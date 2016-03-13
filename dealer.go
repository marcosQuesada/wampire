package wampire

import (
	"fmt"
	"sync"
	"log"
)

type Dealer interface {
	Register(URI, Handler) error
	Call(Message, Peer) Message
}

type Handler func(Message, Peer) (Message, error)

type defaultDealer struct {
	handlers map[URI]Handler
	mutex    *sync.Mutex
}

func NewDealer() *defaultDealer {
	return &defaultDealer{
		handlers: make(map[URI]Handler),
		mutex: &sync.Mutex{},
	}
}

func (d *defaultDealer) Register(uri URI, fn Handler) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if _, ok := d.handlers[uri]; !ok {
		return fmt.Errorf("%s handler already registered ", uri)
	}
	d.handlers[uri] = fn

	return nil
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
	handler, ok := d.handlers[call.Procedure]
	if !ok {
		uri := "Handlers not found"
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
