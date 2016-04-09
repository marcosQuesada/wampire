package core

import (
	"fmt"
	"log"
	"sync"
)

type Dealer interface {
	Register(Message, *Session)
	Unregister(Message, *Session)
	Call(Message, *Session)
	Yield(Message, *Session)
	Cancel(Message, *Session)
	RegisterSessionHandlers(map[URI]Handler, *inSession)
	Handlers() map[URI]Handler
}

type defaultDealer struct {
	sessionHandlers map[URI]ID
	registrations   map[ID]*Session // handler registrations from session callees
	reqListeners    *RequestListener
	mutex           *sync.RWMutex
	metaEvents      SessionMetaEventHandler
	currentTasks    map[ID]chan struct{}
}

func NewDealer(m SessionMetaEventHandler) *defaultDealer {
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

func (d *defaultDealer) Register(msg Message, s *Session) {
	log.Println("Register procedure: ", msg.(*Register).Procedure)
	d.mutex.Lock()
	defer d.mutex.Unlock()

	register := msg.(*Register)
	if _, ok := d.sessionHandlers[register.Procedure]; ok {
		uri := fmt.Sprintf("%s handler already registered ", register.Procedure)
		response := &Error{
			Request: register.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
	}

	id := NewId()
	d.registrations[id] = s
	d.sessionHandlers[register.Procedure] = id
	s.addRegistration(id, register.Procedure)
	d.metaEvents.Fire(
		s.ID(),
		URI("wampire.registration.on_register"),
		map[string]interface{}{},
	)

	if s.ID() == PeerID("internal") {
		return
	}
	response := &Registered{
		Request:      register.Request,
		Registration: id,
	}
	s.Send(response)
}

func (d *defaultDealer) Unregister(msg Message, s *Session) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	unregister := msg.(*Unregister)
	_, ok := d.registrations[unregister.Registration]
	if !ok {
		uri := fmt.Sprintf("%d handler not registered ", unregister.Request)
		log.Println(uri)
		response := &Error{
			Request: unregister.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
	}

	uri, err := s.uriFromRegistration(unregister.Registration)
	if err != nil {
		uri := fmt.Sprintf("%d handler not found on Session ", unregister.Registration)
		log.Println(uri)
		response := &Error{
			Request: unregister.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
	}

	_, ok = d.sessionHandlers[uri]
	if !ok {
		uri := fmt.Sprintf("%d peerHandlers not found  %s", unregister.Registration, uri)
		log.Println(uri)
		response := &Error{
			Request: unregister.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
	}

	//delete uri s
	delete(d.sessionHandlers, uri)
	//delete dealer registrations
	delete(d.registrations, unregister.Registration)
	//unregister uri from session
	s.unregister(uri)

	d.metaEvents.Fire(
		s.ID(),
		URI("wampire.registration.on_unregister"),
		map[string]interface{}{},
	)

	response := &Unregistered{
		Request: unregister.Request,
	}
	s.Send(response)
}

func (d *defaultDealer) Call(msg Message, s *Session) {
	if msg.MsgType() != CALL {
		uri := "Unexpected message type on Call"
		log.Print(uri, msg.MsgType())
		response := &Error{
			Error: URI(uri),
		}
		s.Send(response)
		return
	}

	call := msg.(*Call)
	registration, ok := d.sessionHandlers[call.Procedure]
	if !ok {
		uri := "Registration not found on sessionHandlers"
		log.Print(uri, msg.MsgType())
		response := &Error{
			Error: URI(uri),
		}
		s.Send(response)
		return
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
		response := &Error{
			Error: URI(uri),
		}
		s.Send(response)
		return
	}

	err := calleeSession.do(invocation)
	if err != nil {
		log.Println("Error calleeSession do", err, invocation)
		response := &Error{
			Error: URI("Pending to develop"),
		}
		s.Send(response)
		return
	}

	y, err := d.reqListeners.RegisterAndWait(invocation.Request)
	if err != nil {
		log.Println("Error waiting response ", err, msg)
		response := &Error{
			Error: URI("Pending to develop"),
		}
		s.Send(response)
		return
	}

	yield := y.(*Yield)

	response := &Result{
		Request:     yield.Request,
		Details:     yield.Options,
		Arguments:   yield.Arguments,
		ArgumentsKw: yield.ArgumentsKw,
	}
	s.Send(response)
}

func (d *defaultDealer) Yield(msg Message, s *Session) {
	externalResult := msg.(*Yield)
	d.reqListeners.Notify(msg, externalResult.Request)

	return
}

func (d *defaultDealer) Cancel(msg Message, s *Session) {
	// @TODO: How to cancel a job in progress?
	response := &Error{
		Error: URI("Pending to develop"),
	}

	s.Send(response)
}

func (d *defaultDealer) RegisterSessionHandlers(handlers map[URI]Handler, s *inSession) {
	for uri, h := range handlers {
		d.Register(&Register{Request: NewId(), Procedure: uri}, s.session)
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
