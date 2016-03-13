package wampire

import (
	"fmt"
	"sync"
	"log"
)

type Dealer interface {
	Register(URI, Handler) error
	Call(Message, Peer) Message
	RegisterExternalHandler(Message, Peer) Message
	UnregisterExternalHandler(Message, Peer) Message
}

type Handler func(Message, Peer) (Message, error)

type defaultDealer struct {
	handlers map[URI]Handler
	delegatedHandlers map[ID]URI
	mutex    *sync.RWMutex
}

func NewDealer() *defaultDealer {
	return &defaultDealer{
		handlers: make(map[URI]Handler),
		delegatedHandlers: make(map[ID]URI),
		mutex: &sync.RWMutex{},
	}
}

func (d *defaultDealer) Register(uri URI, fn Handler) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if _, ok := d.handlers[uri]; ok {
		return fmt.Errorf("%s handler already registered ", uri)
	}
	d.handlers[uri] = fn

	return nil
}

func (d *defaultDealer) unregister(uri URI) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if _, ok := d.handlers[uri]; !ok {
		return fmt.Errorf("%s handler not registered ", uri)
	}
	delete(d.handlers, uri)

	return nil
}
func (d *defaultDealer) RegisterExternalHandler(msg Message, p Peer) Message {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	register := msg.(*Register)
	if _, ok := d.handlers[register.Procedure]; ok {
		uri :=fmt.Sprintf("%s handler already registered ", register.Procedure)
		return &Error{
			Request: register.Request,
			Error: URI(uri),
		}
	}
	// Register peer that offers uri service
	delegated := NewDelegatedCall(register.Procedure, p)
	d.handlers[register.Procedure] = delegated.delegatedCall
	d.delegatedHandlers[delegated.id] = register.Procedure

	return &Registered{
		Request: register.Request,
		Registration: delegated.id,
	}
}

func (d *defaultDealer) UnregisterExternalHandler(msg Message, p Peer) Message {
	unregister := msg.(*Unregister)
	uri, ok := d.delegatedHandlers[unregister.Registration]
	if !ok {
		uri :=fmt.Sprintf("%s handler not registered ", unregister.Request)
		return &Error{
			Request: unregister.Request,
			Error: URI(uri),
		}
	}
	err := d.unregister(uri)
	if err!=nil {
		uri :=fmt.Sprintf("Unexpected error unregistering delegated Handler")
		log.Println(uri, err)

		return &Error{
			Request: unregister.Request,
			Error: URI(uri),
		}
	}

	return &Unregistered{
		Request: unregister.Request,
	}
}

func (d *defaultDealer) Call(msg Message, p Peer) Message {
	if msg.MsgType() != CALL {
		uri := "Unexpected message type on Call"
		log.Print(uri, msg.MsgType())
		return &Error{
			Error: URI(uri),
		}
	}
	call := msg.(*Call)
	d.mutex.RLock()
	handler, ok := d.handlers[call.Procedure]
	d.mutex.RUnlock()
	if !ok {
		uri := "Handler not found"
		log.Print(uri, msg.MsgType())
		return &Error{
			Error: URI(uri),
		}
	}

	response, err := handler(call, p)
	if err!= nil {
		uri := "Error invoking Handler"
		log.Print(uri, err)
		return &Error{
			Error: URI(uri),
		}
	}

	return response
}

type delegatedCall struct {
	procedure URI
	peer Peer
	id ID
}

func NewDelegatedCall(uri URI, p Peer) *delegatedCall {
	return &delegatedCall{
		procedure: uri,
		peer: p,
		id: NewId(),
	}
}

func (d *delegatedCall) delegatedCall(msg Message, p Peer) (Message, error){
	rsp := p.Request(msg)
	log.Println("Result is ", rsp.MsgType())
	//How to handle response!
	return &Error{
		Error: URI("Pending to develop"),
	}, nil
}