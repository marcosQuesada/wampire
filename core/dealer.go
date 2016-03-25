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
	RegisterSessionHandlers(*inSession)
}

type defaultDealer struct {
	sessionHandlers map[URI]ID
	registrations   map[ID]*Session  // handler registrations from session callees
	reqListeners    *RequestListener
	//currentTasks: make(map[ID]Message //@TODO: Register active Calls to enable Cancel
	mutex *sync.RWMutex
}

func NewDealer() *defaultDealer {
	d := &defaultDealer{
		sessionHandlers: make(map[URI]ID),
		registrations:   make(map[ID]*Session),
		mutex:        &sync.RWMutex{},
		reqListeners: NewRequestListener(),
	}

	return d
}

func (d *defaultDealer) Register(msg Message, s *Session) Message {
	log.Println("Register invoked ", msg.(*Register))
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

	log.Println("Invocation Request is ", call.Request, "org peer ", s.ID())
	calleeSession, ok := d.registrations[registration]
	if !ok {
		uri := "Registration not foun"
		log.Print(uri, msg.MsgType())
		return &Error{
			Error: URI(uri),
		}
	}


	//@TODO: Key point to improve inernal client performance
	err := calleeSession.do(invocation)
	if err != nil {
		log.Println("Error calleeSession do", err, invocation)
		return &Error{
			Error: URI("Pending to develop"),
		}
	}

	log.Println("Invocation DOne  ", call.Request, "org peer ", s.ID())
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
	log.Println("Yield invoked ", externalResult.Request)
	d.reqListeners.Notify(msg, externalResult.Request)

	return nil
}

func (d *defaultDealer) Cancel(msg Message, p *Session) Message {
	// @TODO: How to cancel a job in progress?
	return &Error{
		Error: URI("Pending to develop"),
	}
}

func (d *defaultDealer) RegisterSessionHandlers(s *inSession) {
	for uri, _ := range s.Handlers() {
		msg := d.Register(&Register{Request: NewId(), Procedure: uri}, s.session)
		if msg.MsgType() != REGISTERED {
			log.Println("InSession Error registerig ", uri)
		}

		// Registration!
		r := msg.(*Registered)
		s.session.addRegistration(r.Registration, uri)
	}
}