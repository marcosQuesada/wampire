package core

import (
	"testing"
)

func TestInSessionHandlers(t *testing.T) {
	s := newInSession()
	registerHandlers(s)
	invocation := &Invocation{
		Request: ID(1234),
	}
	yield, err := s.list(invocation)
	if err != nil {
		t.Error("Error on listSessions invocation ", err)
	}
	if len(yield.(*Yield).Arguments) != 3 {
		t.Error("Unexpected Yield arguments result ", yield.(*Yield).Arguments)
	}

	//Avoid failing, result may come disordered
	for _, v := range yield.(*Yield).Arguments {
		if v.(URI) == URI("wampire.core.help") {
			continue
		}
		if v.(URI) == URI("wampire.core.list") {
			continue
		}
		if v.(URI) == URI("wampire.core.echo") {
			continue
		}
		t.Error("Unexpected Yield arguments result ", yield.(*Yield).Arguments, v)
	}
}

func registerHandlers(s *inSession) {
	//Register insession handlers
	for uri, handler := range s.Handlers() {
		err := s.session.register(uri, handler)
		if err != nil {
			panic("Unexpected error registering handlers")
		}

	}

}
