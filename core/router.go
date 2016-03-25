package core

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Router struct {
	sessions map[PeerID]*Session
	broker   Broker
	dealer   Dealer
	exit     chan struct{}
	mutex    *sync.RWMutex
	auth     Authenticator
}

type Authenticator func(Message) bool

func NewRouter() *Router {

	return &Router{
		sessions: make(map[PeerID]*Session),
		broker:   NewBroker(),
		dealer:   NewDealer(),
		exit:     make(chan struct{}),
		mutex:    &sync.RWMutex{},
	}


}

func (r *Router) Accept(p Peer) error {
	timeout := time.NewTimer(time.Second * 1)
	select {
	case rcvMessage := <-p.Receive():
		timeout.Stop()
		h, ok := rcvMessage.(*Hello)
		if !ok {
			err := fmt.Sprintf("Unexpected type on Accept: %d", rcvMessage.MsgType())
			log.Println(err)

			return fmt.Errorf(err)
		}

		var response Message
		if r.authenticate(h) {
			response = &Abort{
				//Id: h.Id,
			}
		}

		//answer welcome
		response = &Welcome{
			Id: NewId(),
			Details: h.Details,
		}
		p.Send(response)

		session := NewSession(p)
		r.register(session)

		go r.handleSession(session)

		return nil
	case <-timeout.C:
		log.Println("Timeout error waiting Hello Message")
		return fmt.Errorf("Timeout error waiting Hello Message")
	}
}

func (r *Router) Terminate() {
	close(r.exit)

	//wait until all handleSession has finished
	<-r.waitUntilVoid()
	log.Println("Router terminated!")
}

func (r *Router) SetAuthenticator(a Authenticator) {
	r.auth = a
}

func (r *Router) authenticate(msg Message) bool {
	if r.auth != nil {
		return r.auth(msg)
	}

	return true
}

func (r *Router) handleSession(s *Session) {
	defer log.Println("Exit session handler from peer ", s.ID())
	defer s.Terminate()
	defer r.unRegister(s)
	defer func() {
		for sid, topic := range s.subscriptions {
			log.Printf("Unsubscribe sid %d on topic %s \n", sid, topic)
			u := &Unsubscribe{Request: NewId(), Subscription: sid}
			r.broker.UnSubscribe(u, s)
		}
	}()

	for {
		select {
		case msg, open := <-s.Receive():
			if !open {
				log.Println("Closing handled session from closed receive chan")
				return
			}

			var response Message
			switch msg.(type) {
			case *Goodbye:
				log.Println("Received Goodbye")
				// exit handler
			case *Publish:
				log.Println("Received Publish")
				response = r.broker.Publish(msg, s)
			case *Subscribe:
				log.Println("Received Subscribe")
				response = r.broker.Subscribe(msg, s)

				//store subscription on session
				//unsubscribe on session close
				if sbd, ok := response.(*Subscribed); ok {
					r.mutex.Lock()
					s.subscriptions[sbd.Subscription] = msg.(*Subscribe).Topic
					r.mutex.Unlock()
				}
			case *Unsubscribe:
				log.Println("Received Unubscribe")
				response = r.broker.UnSubscribe(msg, s)

				// remove subscription from session
				if _, ok := response.(*Unsubscribed); ok {
					usbd := msg.(*Unsubscribe)
					r.mutex.Lock()
					delete(s.subscriptions, usbd.Subscription)
					r.mutex.Unlock()
				}
			case *Call:
				log.Println("Received Call")
				response = r.dealer.Call(msg, s)
			case *Cancel:
				log.Println("Received Cancel, forward this to requester on dealer ")
				r.dealer.Cancel(msg, s)

			case *Yield:
				log.Println("Received Yield, forward this to dealer ")
				r.dealer.Yield(msg, s)

			//@TODO: Communication between wamp nodes
			case *Invocation:
				log.Println("Received Invocation, execute on callee")
			//	r.dealer.Invocation(msg, p)
			case *Result:
				log.Println("Received Result, Unexpected message on wamp router")
				//r.dealer.ExternalResult(msg, p)

			case *Register:
				log.Println("Received Register")
				response = r.dealer.Register(msg, s)
			case *Unregister:
				log.Println("Received Unregister")
				response = r.dealer.Unregister(msg, s)
			default:
				log.Println("Session unhandled message ", msg.MsgType())
				response = &Error{
					Error: URI(fmt.Sprintf("Session unhandled message &d", msg.MsgType())),
				}
			}
			if response != nil {
				s.Send(response)
			}
		case <-r.exit:
			log.Println("Shutting down session handler from peer ", s.ID())
			return
		}
	}
}

func (r *Router) register(p *Session) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.sessions[p.ID()]; ok {
		return fmt.Errorf("Peer %s already registered", p.ID())
	}

	r.sessions[p.ID()] = p

	return nil
}

func (r *Router) unRegister(p *Session) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.sessions[p.ID()]; !ok {
		return fmt.Errorf("Peer %s not registered", p.ID())
	}

	delete(r.sessions, p.ID())

	return nil
}

// waitUntilVoid: waits until all sessions are closed
func (r *Router) waitUntilVoid() chan struct{} {
	void := make(chan struct{})
	go func() {
		for {
			r.mutex.RLock()
			sessions := len(r.sessions)
			r.mutex.RUnlock()
			if sessions == 0 {
				close(void)
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	return void
}
