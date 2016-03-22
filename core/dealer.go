package core

import (
	"fmt"
	"log"
	"sync"
)

type Dealer interface {
	Register(URI, Handler) error
	Call(Message, Peer) Message
	RegisterExternalHandler(Message, Peer) Message
	UnregisterExternalHandler(Message, Peer) Message
	ExternalResult(Message, Peer)
}

type Handler func(Message, Peer) (Message, error)

type defaultDealer struct {
	handlers          map[URI]Handler
	delegatedHandlers map[ID]URI
	mutex             *sync.RWMutex
	reqListeners      *RequestListener
}

func NewDealer() *defaultDealer {
	return &defaultDealer{
		handlers:          make(map[URI]Handler),
		delegatedHandlers: make(map[ID]URI),
		mutex:             &sync.RWMutex{},
		reqListeners:      NewRequestListener(),
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
		uri := fmt.Sprintf("%s handler already registered ", register.Procedure)
		return &Error{
			Request: register.Request,
			Error:   URI(uri),
		}
	}
	// Register peer that offers uri service
	delegated := d.delegateCall(register.Procedure, p)
	d.handlers[register.Procedure] = delegated.externalCall
	d.delegatedHandlers[delegated.id] = register.Procedure

	return &Registered{
		Request:      register.Request,
		Registration: delegated.id,
	}
}

func (d *defaultDealer) UnregisterExternalHandler(msg Message, p Peer) Message {
	unregister := msg.(*Unregister)
	uri, ok := d.delegatedHandlers[unregister.Registration]
	if !ok {
		uri := fmt.Sprintf("%s handler not registered ", unregister.Request)
		return &Error{
			Request: unregister.Request,
			Error:   URI(uri),
		}
	}
	err := d.unregister(uri)
	if err != nil {
		uri := fmt.Sprintf("Unexpected error unregistering delegated Handler")
		log.Println(uri, err)

		return &Error{
			Request: unregister.Request,
			Error:   URI(uri),
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
	if err != nil {
		uri := "Error invoking Handler"
		log.Print(uri, err)
		return &Error{
			Error: URI(uri),
		}
	}

	return response
}
func (d *defaultDealer) ExternalResult(msg Message, p Peer) {
	//@TODO: SURE TO HANDLE THIS FROM REQUEST LISTENERS ¿?
	externalResult := msg.(*Result)
	d.reqListeners.Notify(msg, externalResult.Request)
}

type delegatedCall struct {
	reqListener *RequestListener //@TODO: Rethink this
	procedure   URI
	peer        Peer
	id          ID
}

func (d *defaultDealer) delegateCall(uri URI, p Peer) *delegatedCall {
	return &delegatedCall{
		procedure:   uri,
		peer:        p,
		id:          NewId(),
		reqListener: d.reqListeners,
	}
}

func (d *delegatedCall) externalCall(msg Message, p Peer) (Message, error) {
	call := msg.(*Call)
	log.Println("Delegated Call Request is ", call.Request, "dest peer ", d.peer.ID())
	d.peer.Send(call)

	//How to handle response!
	msg, err := d.reqListener.RegisterAndWait(call.Request)
	if err != nil {
		log.Println("Error waiting response ", err, msg)
		return &Error{
			Error: URI("Pending to develop"),
		}, nil
	}

	return msg, nil
}
