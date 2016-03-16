package wampire

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Router struct {
	peers  map[ID]Peer
	broker Broker
	dealer Dealer
	exit   chan struct{}
	wg     *sync.WaitGroup
}

func NewRouter() *Router {
	return &Router{
		peers:  make(map[ID]Peer),
		broker: NewBroker(),
		dealer: NewDealer(),
		exit:   make(chan struct{}),
		wg:     &sync.WaitGroup{},
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
				Id: h.Id,
			}
		}

		//answer welcome
		response = &Welcome{
			Id: h.Id,
		}
		p.Send(response)
		//if all goes fine
		session := NewSession(p)
		go r.handleSession(session)

		return nil
	case <-timeout.C:
		log.Println("Timeout error waiting Hello Message")
		return fmt.Errorf("Timeout error waiting Hello Message")
	}
}

func (r *Router) Terminate() {
	log.Println("Invoked Router terminated!")
	close(r.exit)
	r.wg.Wait()
	log.Println("Router terminated!")
}

func (r *Router) handleSession(p *Session) {
	defer log.Println("Exit session handler from peer ", p.ID())
	defer r.wg.Done()
	defer p.Terminate()
	defer func() {
		for sid, topic := range p.subscriptions {
			log.Println("Unsubscribe sid %s on topic %s", sid, topic)
			u := &Unsubscribe{Request: NewId(), Subscription: sid}
			r.broker.UnSubscribe(u, nil)
		}
	}()
	r.wg.Add(1)

	for {
		select {
		case msg, open := <-p.Receive():
			if !open {
				log.Println("Closing handled session from closed receive chan")
				return
			}

			var response Message
			switch msg.(type) {
			case *Publish:
				log.Println("Received Publish")
				response = r.broker.Publish(msg, p)
			case *Subscribe:
				log.Println("Received Subscribe")
				response = r.broker.Subscribe(msg, p)
				//store subscription on session
				//unsubscribe on session close
				if s, ok := response.(*Subscribed); ok {
					p.subscriptions[s.Subscription] = msg.(*Subscribe).Topic
				}
			case *Unsubscribe:
				log.Println("Received Subscribe")
				response = r.broker.UnSubscribe(msg, p)
				// remove subscription from session
				if _, ok := response.(*Unsubscribed); ok {
					s := msg.(*Unsubscribe)
					delete(p.subscriptions, s.Subscription)
				}
			case *Call:
				log.Println("Received Call")
				response = r.dealer.Call(msg, p)

			// First approach on remote Handlers, used as result callback
			case *Register:
				log.Println("Received Register")
				response = r.dealer.RegisterExternalHandler(msg, p)
			case *Unregister:
				log.Println("Received Unregister")
				response = r.dealer.UnregisterExternalHandler(msg, p)
			case *Result:
				log.Println("Received External Result, forward this to requester on dealer ")
				r.dealer.ExternalResult(msg, p)
			default:
				log.Println("Session unhandled message ", msg.MsgType())
				response = &Error{
					Error: URI(fmt.Sprintf("Session unhandled message &d", msg.MsgType())),
				}
			}
			if response != nil {
				p.Send(response)
			}
		case <-r.exit:
			log.Println("Shutting down session handler from peer ", p.ID())
			return
		}
	}
}

func (r *Router) authenticate(msg Message) bool {
	return true
}
