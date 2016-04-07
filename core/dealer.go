package core

import (
	"fmt"
	"log"
	"sync"
)

type Dealer interface {
	Register(Message, *Session) Message
	Unregister(Message, *Session) Message
	Call(Message, *Session) Message
	Yield(Message, *Session) Message
	Cancel(Message, *Session) Message
	RegisterSessionHandlers(map[URI]Handler, *inSession)
	Handlers() map[URI]Handler
}

type defaultDealer struct {
	sessionHandlers map[URI]ID
	registrations   map[ID]*Session // handler registrations from session callees
	reqListeners    *RequestListener
	mutex           *sync.RWMutex
	metaEvents      *SessionMetaEventHandler
	currentTasks    map[ID]chan struct{}
}

func NewDealer(m *SessionMetaEventHandler) *defaultDealer {
	d := &defaultDealer{
		sessionHandlers: make(map[URI]ID),
		registrations:   make(map[ID]*Session),
		mutex:           &sync.RWMutex{},
		reqListeners:    NewRequestListener(),
		metaEvents:      m,
		currentTasks:    make(map[ID]chan struct{}),
	}

	return d
}

func (d *defaultDealer) Register(msg Message, s *Session) Message {
	log.Println("Register invoked ", msg.(*Register).Procedure)
	d.mutex.Lock()
	defer d.mutex.Unlock()
	register := msg.(*Register)
	if _, ok := d.sessionHandlers[register.Procedure]; ok {
		uri := fmt.Sprintf("%s handler already registered ", register.Procedure)
		return &Error{
			Request: register.Request,
			Error:   URI(uri),
		}
	}

	id := NewId()
	d.registrations[id] = s
	d.sessionHandlers[register.Procedure] = id
	s.addRegistration(id, register.Procedure)
	d.metaEvents.fireMetaEvents(
		s.ID(),
		URI("wampire.registration.on_register"),
		map[string]interface{}{},
	)

	return &Registered{
		Request:      register.Request,
		Registration: id,
	}
}

func (d *defaultDealer) Unregister(msg Message, s *Session) Message {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	unregister := msg.(*Unregister)
	_, ok := d.registrations[unregister.Registration]
	if !ok {
		uri := fmt.Sprintf("%d handler not registered ", unregister.Request)
		log.Println(uri)
		return &Error{
			Request: unregister.Request,
			Error:   URI(uri),
		}
	}

	uri, err := s.uriFromRegistration(unregister.Registration)
	if err != nil {
		uri := fmt.Sprintf("%d handler not found on Session ", unregister.Registration)
		log.Println(uri)
		return &Error{
			Request: unregister.Request,
			Error:   URI(uri),
		}
	}

	_, ok = d.sessionHandlers[uri]
	if !ok {
		uri := fmt.Sprintf("%d peerHandlers not found  %s", unregister.Registration, uri)
		log.Println(uri)
		return &Error{
			Request: unregister.Request,
			Error:   URI(uri),
		}
	}

	//delete uri s
	delete(d.sessionHandlers, uri)
	//delete dealer registrations
	delete(d.registrations, unregister.Registration)
	//unregister uri from session
	s.unregister(uri)

	d.metaEvents.fireMetaEvents(
		s.ID(),
		URI("wampire.registration.on_unregister"),
		map[string]interface{}{},
	)

	return &Unregistered{
		Request: unregister.Request,
	}
}

func (d *defaultDealer) Call(msg Message, s *Session) Message {
	if msg.MsgType() != CALL {
		uri := "Unexpected message type on Call"
		log.Print(uri, msg.MsgType())
		return &Error{
			Error: URI(uri),
		}
	}

	call := msg.(*Call)
	registration, ok := d.sessionHandlers[call.Procedure]
	if !ok {
		uri := "Registration not found on sessionHandlers"
		log.Print(uri, msg.MsgType())
		return &Error{
			Error: URI(uri),
		}
	}
	// Forward as invocation to peer calleee
	invocation := &Invocation{
		Request:      call.Request,
		Registration: registration,
		Details:      call.Options,
		Arguments:    call.Arguments,
		ArgumentsKw:  call.ArgumentsKw,
	}

	log.Println("Invocation Request is ", call.Request, "origin peer ", s.ID())
	calleeSession, ok := d.registrations[registration]
	if !ok {
		uri := "Registration not found "
		log.Print(uri, msg.MsgType())
		return &Error{
			Error: URI(uri),
		}
	}

	err := calleeSession.do(invocation)
	if err != nil {
		log.Println("Error calleeSession do", err, invocation)
		return &Error{
			Error: URI("Pending to develop"),
		}
	}

	y, err := d.reqListeners.RegisterAndWait(invocation.Request)
	if err != nil {
		log.Println("Error waiting response ", err, msg)
		return &Error{
			Error: URI("Pending to develop"),
		}
	}

	yield := y.(*Yield)

	return &Result{
		Request:     yield.Request,
		Details:     yield.Options,
		Arguments:   yield.Arguments,
		ArgumentsKw: yield.ArgumentsKw,
	}
}

func (d *defaultDealer) Yield(msg Message, p *Session) Message {
	externalResult := msg.(*Yield)
	d.reqListeners.Notify(msg, externalResult.Request)

	return nil
}

func (d *defaultDealer) Cancel(msg Message, p *Session) Message {
	// @TODO: How to cancel a job in progress?
	return &Error{
		Error: URI("Pending to develop"),
	}
}

func (d *defaultDealer) RegisterSessionHandlers(handlers map[URI]Handler, s *inSession) {
	for uri, h := range handlers {
		msg := d.Register(&Register{Request: NewId(), Procedure: uri}, s.session)
		if msg.MsgType() != REGISTERED {
			log.Println("InSession Error registerig ", uri)
		}
		err := s.session.register(uri, h)
		if err != nil {
			log.Println("InSession Error registerig ", uri)
		}
	}
}

func (d *defaultDealer) Handlers() map[URI]Handler {
	return map[URI]Handler{
		"wampire.core.dealer.dump": d.dumpDealer,
	}
}

func (d *defaultDealer) dumpDealer(msg Message) (Message, error) {
	d.mutex.RLock()
	sessions := d.sessionHandlers
	registrations := d.registrations
	d.mutex.RUnlock()

	list := map[string]interface{}{}
	for uri, id := range sessions {
		list[string(uri)] = id
	}

	regs := map[string]interface{}{}
	for id, s := range registrations {
		regs[fmt.Sprintf("%d", id)] = s.ID()
	}
	inv := msg.(*Invocation)
	kw := map[string]interface{}{
		"session_handlers": list,
		"registrations":    regs,
	}

	return &Yield{
		Request:     inv.Request,
		ArgumentsKw: kw,
	}, nil

}
