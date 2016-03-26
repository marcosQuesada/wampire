package core

import "log"

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

func (i *inSession) Handlers() map[URI]Handler {
	return map[URI]Handler{
		"wampire.core.help": i.help,
		"wampire.core.list": i.list,
	}
}

func (i *inSession) readLoop() {
	for {
		select {
		case msg := <-i.session.Receive():
			log.Println("internal client session receive ", msg.MsgType(), i.session.ID())
		case <-i.done:
			return
		}
	}
}

func (i *inSession) exit() {
	close(i.done)
}
