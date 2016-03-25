package core

import "log"

type inSession struct {
	session *Session
	done chan struct{}
}

func newInSession() *inSession {
	internalSession := NewSession(NewInternalPeer())
	i := &inSession{
		session: internalSession,
		done: make(chan struct{}),
	}
	//go i.readLoop()

	for u, h := range i.Handlers() {
		err := internalSession.register(u, h)
		if err != nil {
			log.Println("InSession Error registerig ", u)
		}
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
	inv := msg.(*Invocation)
	return &Yield{
		Request:   inv.Request,
		Arguments: []interface{}{"list-okiDoki"},
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