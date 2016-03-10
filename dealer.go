package wamp

import "fmt"

type Dealer interface {
	Register(URI, Handler) error
	Call(Message) Message
}

type Handler func(Peer, Message) (Message, error)

type defaultDealer struct{
	handlers map[URI]Handler
}

func NewDealer() *defaultDealer {
	return &defaultDealer{
		handlers: make(map[URI]Handler),
	}
}

func (d *defaultDealer) Register(uri URI, fn Handler) error {
	if _, ok:= d.handlers[uri]; !ok {
		return fmt.Errorf("%s handler already registered ", uri)
	}
	d.handlers[uri] = fn

	return nil
}


func (d *defaultDealer) Call(Message) Message {
	return &Result{

	}
}