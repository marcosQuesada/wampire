package core

import (
	"log"
	"github.com/socialpoint/sprocket/pkg/dumper"
)

type inSession struct {
	session *Session
	done    chan struct{}
}

func newInSession() *inSession {
	internalSession := NewSession(NewInternalPeer())
	i := &inSession{
		session: internalSession,
		done:    make(chan struct{}),
	}
	//go i.readLoop()

	return i
}

func (i *inSession) help(msg Message) (Message, error) {
	inv := msg.(*Invocation)

	return &Yield{
		Request:   inv.Request,
		Arguments: []interface{}{"help-okiDoki"},
	}, nil
}

func (i *inSession) list(msg Message) (Message, error) {
	list := []interface{}{}
	for uri, _ := range i.session.handlers {
		list = append(list, uri)
	}

	inv := msg.(*Invocation)
	return &Yield{
		Request:   inv.Request,
		Arguments: list,
	}, nil
}

func (i *inSession) echo(msg Message) (Message, error) {
	inv := msg.(*Invocation)
	res := append(inv.Arguments, "oki")

	return &Yield{
		Request:   inv.Request,
		Arguments: res,
	}, nil
}

func (i *inSession) Handlers() map[URI]Handler {
	return map[URI]Handler{
		"wampire.core.help": i.help,
		"wampire.core.list": i.list,
		"wampire.core.echo": i.echo,
	}
}

func (i *inSession) readLoop() {
	for {
		select {
		case msg := <-i.session.Receive():
			log.Println(msg.MsgType())
			dumper.B(msg)
			if err, ok := msg.(*Error); ok {
				log.Println("Received error is ", err.Error)
			}
		case <-i.done:
			return
		}
	}
}

func (i *inSession) exit() {
	close(i.done)
}
