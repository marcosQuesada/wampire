package core

type inSession struct {
	session *Session
}

func newInSession() *inSession {
	return &inSession{
		session: newInternalSession(),
	}
}

func newInternalSession() *Session {
	return NewSession(NewInternalPeer())
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
