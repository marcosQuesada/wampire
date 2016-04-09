package core

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Router interface {
	Accept(p Peer) error
	Terminate()
	SetAuthenticator(a Authenticator)
}

type DefaultRouter struct {
	sessions map[PeerID]*Session
	Broker
	Dealer
	exit            chan struct{}
	mutex           *sync.RWMutex
	auth            Authenticator
	internalSession *inSession
	metaEvents      SessionMetaEventHandler
}

type Authenticator func(Message) bool

func NewRouter() *DefaultRouter {
	internalSession := newInSession()
	m := NewSessionMetaEventsHandler()
	r := &DefaultRouter{
		sessions:        make(map[PeerID]*Session),
		Broker:          NewBroker(m),
		Dealer:          NewDealer(m),
		exit:            make(chan struct{}),
		mutex:           &sync.RWMutex{},
		internalSession: internalSession,
		metaEvents:      m,
	}

	// Handle Session Meta Events
	go m.Consume(r)

	// Register in session procedures
	r.Dealer.RegisterSessionHandlers(internalSession.Handlers(), internalSession)
	r.Dealer.RegisterSessionHandlers(r.Handlers(), internalSession)
	r.Dealer.RegisterSessionHandlers(r.Broker.Handlers(), internalSession)
	r.Dealer.RegisterSessionHandlers(r.Dealer.Handlers(), internalSession)

	//Handle internal Session
	go r.handleSession(internalSession.session)

	return r
}

func (r *DefaultRouter) Accept(p Peer) error {
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

func (r *DefaultRouter) Terminate() {
	close(r.exit)

	//wait until all handleSession has finished
	<-r.waitUntilVoid()
	log.Println("Router terminated!")
	r.metaEvents.Terminate()
}

func (r *DefaultRouter) SetAuthenticator(a Authenticator) {
	r.auth = a
}

func (r *DefaultRouter) authenticate(msg Message) (Message, bool, error) {
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

func (r *DefaultRouter) handleSession(s *Session) {
	defer func() {
		log.Println("Exit session handler from peer ", s.ID())
		// remove session subscriptions
		for sid, topic := range s.subscriptions {
			log.Printf("Unsubscribe sid %d on topic %s \n", sid, topic)
			u := &Unsubscribe{Request: NewId(), Subscription: sid}
			r.Broker.UnSubscribe(u, s)
		}
		// Fire on_leave Session Meta Event
		r.metaEvents.Fire(s.ID(), URI("wampire.session.on_leave"), map[string]interface{}{})
		//unregister session from router
		r.unRegister(s)
		//exit session
		s.Terminate()
	}()

	//Fire on_join Session Meta Event
	r.metaEvents.Fire(s.ID(), URI("wampire.session.on_join"), map[string]interface{}{})

	for {
		select {
		case msg, open := <-s.Receive():
			if !open {
				log.Println("Closing handleSession from closed receive chan")
				return
			}

			switch msg.(type) {
			case *Goodbye:
				log.Println("Received Goodbye, exit handle session")
				return
			case *Publish:
				log.Println("Received Publish on topic ", msg.(*Publish).Topic)
				go r.Broker.Publish(msg, s)
			case *Subscribe:
				log.Println("Received Subscribe", msg.(*Subscribe).Topic)
				go r.Broker.Subscribe(msg, s)
			case *Unsubscribe:
				log.Println("Received Unubscribe")
				go r.Broker.UnSubscribe(msg, s)
			case *Call:
				log.Println("Received Call ", msg.(*Call).Procedure)
				go r.Dealer.Call(msg, s)
			case *Cancel:
				log.Println("Received Cancel, forward this to requester on dealer ")
				go r.Dealer.Cancel(msg, s)
			case *Yield:
				log.Println("Received Yield, forward this to dealer ", msg.(*Yield))
				go r.Dealer.Yield(msg, s)
			case *Register:
				log.Println("Received Register ", msg.(*Register).Procedure)
				go r.Dealer.Register(msg, s)
			case *Unregister:
				log.Println("Received Unregister")
				go r.Dealer.Unregister(msg, s)

			//[WIP] Unexpected messages
			case *Invocation:
				log.Println("Received Invocation, execute on callee")
			case *Result:
				log.Println("Received Result, Unexpected message on wamp router")
			case *Published:
				// response from Internal Peer DO Nothing
			default:
				log.Println("Session unhandled message ", msg.MsgType())
			}
		case <-r.exit:
			log.Println("Shutting down session handler from peer ", s.ID())
			return
		}
	}
}

func (r *DefaultRouter) register(p *Session) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.sessions[p.ID()]; ok {
		return fmt.Errorf("Peer %s already registered", p.ID())
	}

	r.sessions[p.ID()] = p

	return nil
}

func (r *DefaultRouter) unRegister(p *Session) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.sessions[p.ID()]; !ok {
		return fmt.Errorf("Peer %s not registered", p.ID())
	}

	delete(r.sessions, p.ID())

	return nil
}

// waitUntilVoid: waits until all sessions are closed
func (r *DefaultRouter) waitUntilVoid() chan struct{} {
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

func (r *DefaultRouter) defaultDetails() map[string]interface{} {
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
