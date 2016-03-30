package core

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Router struct {
	sessions map[PeerID]*Session
	Broker
	Dealer
	exit            chan struct{}
	mutex           *sync.RWMutex
	auth            Authenticator
	internalSession *inSession
}

type Authenticator func(Message) bool

func NewRouter() *Router {
	internalSession := newInSession()
	r := &Router{
		sessions:        make(map[PeerID]*Session),
		Broker:          NewBroker(),
		Dealer:          NewDealer(),
		exit:            make(chan struct{}),
		mutex:           &sync.RWMutex{},
		internalSession: internalSession,
	}

	// Register in session procedures
	r.Dealer.RegisterSessionHandlers(internalSession.Handlers(), internalSession)
	r.Dealer.RegisterSessionHandlers(r.Handlers(), internalSession)
	r.Dealer.RegisterSessionHandlers(r.Broker.Handlers(), internalSession)
	r.Dealer.RegisterSessionHandlers(r.Dealer.Handlers(), internalSession)

	//Handle internal Session
	go r.handleSession(internalSession.session)

	return r
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

		response, auth, err := r.authenticate(h)
		if err != nil {
			log.Println("Error authenticating")
		}
		p.Send(response)
		if !auth {
			log.Println("Authorization denegated, abort")

			return nil
		}

		session := NewSession(p)
		r.register(session)

		go r.handleSession(session)

		return nil
	case <-timeout.C:
		errMsg := "Timeout error waiting Hello Message"
		log.Println(errMsg)
		return fmt.Errorf(errMsg)
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

func (r *Router) authenticate(msg Message) (Message, bool, error) {
	var auth bool = true
	if r.auth != nil {
		auth = r.auth(msg)
	}
	if !auth {
		// @TODO: needs real response
		return &Abort{
			Details: map[string]interface{}{"message": "The realm does not exist."},
			Reason:  URI("wamp.error.no_such_realm"),
		}, false, nil
	}

	return &Welcome{
		Id:      NewId(),
		Details: r.defaultDetails(),
	}, true, nil
}

func (r *Router) handleSession(s *Session) {
	defer func() {
		log.Println("Exit session handler from peer ", s.ID())
		// remove session subscriptions
		for sid, topic := range s.subscriptions {
			log.Printf("Unsubscribe sid %d on topic %s \n", sid, topic)
			u := &Unsubscribe{Request: NewId(), Subscription: sid}
			r.Broker.UnSubscribe(u, s)
		}
		// Fire on_leave Session Meta Event
		r.fireSessionMetaEvent(Topic("wampire.session.meta.events"), s.ID(), "wampire.session.on_leave")
		//unregister session from router
		r.unRegister(s)
		//exit session
		s.Terminate()
	}()

	//Fire on_join Session Meta Event
	r.fireSessionMetaEvent(Topic("wampire.session.meta.events"), s.ID(), "wampire.session.on_join")

	for {
		select {
		case msg, open := <-s.Receive():
			if !open {
				log.Println("Closing handleSession from closed receive chan")
				return
			}

			var response Message
			switch msg.(type) {
			case *Goodbye:
				log.Println("Received Goodbye, exit handle session")
				return
			case *Publish:
				log.Println("Received Publish on topic ", msg.(*Publish).Topic)
				response = r.Broker.Publish(msg, s)
			case *Subscribe:
				log.Println("Received Subscribe", msg.(*Subscribe).Topic)
				response = r.Broker.Subscribe(msg, s)
			case *Unsubscribe:
				log.Println("Received Unubscribe")
				response = r.Broker.UnSubscribe(msg, s)
			case *Call:
				log.Println("Received Call ", msg.(*Call).Procedure)
				response = r.Dealer.Call(msg, s)
			case *Cancel:
				log.Println("Received Cancel, forward this to requester on dealer ")
				r.Dealer.Cancel(msg, s)
			case *Yield:
				log.Println("Received Yield, forward this to dealer ", msg.(*Yield))
				r.Dealer.Yield(msg, s)

			//@TODO: Communication between wamp nodes
			case *Invocation:
				log.Println("Received Invocation, execute on callee")
			//	r.dealer.Invocation(msg, p)
			case *Result:
				log.Println("Received Result, Unexpected message on wamp router")

			case *Register:
				log.Println("Received Register ", msg.(*Register).Procedure)
				response = r.Dealer.Register(msg, s)
			case *Unregister:
				log.Println("Received Unregister")
				response = r.Dealer.Unregister(msg, s)
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

func (r *Router) fireSessionMetaEvent(topic Topic, id PeerID, content string) {
	//Fire on_join Session Meta Event only if is not the internal peer
	if id == PeerID("internal") {
		return
	}
	r.Broker.Publish(
		&Publish{
			Request: NewId(),
			Topic:   topic,
			Options: map[string]interface{}{
				"session_id":  id,
				"acknowledge": true,
			},
			Arguments: []interface{}{
				map[string]interface{}{"message": content},
			},
		}, r.internalSession.session)
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

func (r *Router) Handlers() map[URI]Handler {
	return map[URI]Handler{
		"wampire.session.list":  r.listSessions,
		"wampire.session.count": r.countSessions,
		"wampire.session.get":   r.getSession,
	}
}

func (r *Router) listSessions(msg Message) (Message, error) {
	r.mutex.RLock()
	sessions := r.sessions
	r.mutex.RUnlock()
	list := []interface{}{}
	for peerId, _ := range sessions {
		list = append(list, peerId)
	}

	inv := msg.(*Invocation)

	return &Yield{
		Request:   inv.Request,
		Arguments: list,
	}, nil

}

func (r *Router) countSessions(msg Message) (Message, error) {
	r.mutex.RLock()
	sessions := r.sessions
	r.mutex.RUnlock()
	inv := msg.(*Invocation)

	return &Yield{
		Request:   inv.Request,
		Arguments: []interface{}{len(sessions)},
	}, nil
}

func (r *Router) getSession(msg Message) (Message, error) {
	r.mutex.RLock()
	sessions := r.sessions
	r.mutex.RUnlock()
	inv := msg.(*Invocation)
	log.Println("Getting session ", inv.Arguments)
	if len(inv.Arguments) < 1 {
		error := "Void ID argument on get session"
		log.Println(error)
		return nil, fmt.Errorf(error)
	}

	s, ok := sessions[PeerID(inv.Arguments[0].(string))]
	log.Println("Getting session ", inv.Arguments, "ok:", ok, s)
	if !ok {
		error := "Router session ID %d not found "
		log.Println(error)
		return nil, fmt.Errorf(error)
	}

	subs := map[string]interface{}{}
	for id, topic := range s.getSubscriptions() {
		subs[fmt.Sprintf("%d", id)] = topic
	}
	regs := map[string]interface{}{}
	for id, uri := range s.getRegistrations() {
		regs[fmt.Sprintf("%d", id)] = uri
	}
	kw := map[string]interface{}{
		"subscriptions": subs,
		"registrations": regs,
		"initTs":        s.initTs,
	}

	return &Yield{
		Request:     inv.Request,
		ArgumentsKw: kw,
	}, nil
}

func (r *Router) defaultDetails() map[string]interface{} {
	return map[string]interface{}{
		"roles": map[string]interface{}{
			"publisher": map[string]interface{}{
				"features": map[string]interface{}{
					"publisher_identification":      true,
					"subscriber_blackwhite_listing": true,
					"publisher_exclusion":           true,
				},
			},
			"subscriber": map[string]interface{}{
				"features": map[string]interface{}{
					"publisher_identification": true,
					//"publication_trustlevels": true,
					"pattern_based_subscription": true,
					"subscription_revocation":    true,
					//"event_history": true,
				},
			},
			"broker": map[string]interface{}{
				"features": map[string]interface{}{
					"publisher_identification": true,
					/*					"pattern_based_subscription": true,
										"subscription_meta_api": true,
										"subscription_revocation": true,
										"publisher_exclusion": true,
										"subscriber_blackwhite_listing": true,*/
				},
			},
			"dealer": map[string]interface{}{
				"features": map[string]interface{}{
					"caller_identification": true,
					/*					"progressive_call_results": true,
										"pattern_based_registration": true,
										"registration_revocation": true,
										"shared_registration": true,
										"registration_meta_api": true,*/
				},
			},
			"caller": map[string]interface{}{
				"features": map[string]interface{}{
					"caller_identification": true,
					//"call_timeout": true,
					//"call_canceling": true,
					"progressive_call_results": true,
				},
			},
			"callee": map[string]interface{}{
				"features": map[string]interface{}{
					"caller_identification": true,
					//"call_trustlevels": true,
					"pattern_based_registration": true,
					"shared_registration":        true,
					//"call_timeout": true,
					//"call_canceling": true,
					"progressive_call_results": true,
					"registration_revocation":  true,
				},
			},
		},
	}
}
