package core

import (
	"fmt"
	"sync"
	"testing"
)

var testRouter *DefaultRouter

func TestListSessions(t *testing.T) {
	registerSessions()
	invocation := &Invocation{
		Request: ID(1234),
	}
	yield, err := testRouter.listSessions(invocation)
	if err != nil {
		t.Error("Error on listSessions invocation ", err)
	}
	if len(yield.(*Yield).Arguments) != 3 {
		t.Error("Unexpected Yield arguments result ", yield.(*Yield).Arguments)
	}

	//Avoid failing, result may come disordered
	for _, v := range yield.(*Yield).Arguments {
		if v.(PeerID) == PeerID("fakeSession_0") {
			continue
		}
		if v.(PeerID) == PeerID("fakeSession_1") {
			continue
		}
		if v.(PeerID) == PeerID("fakeSession_2") {
			continue
		}
		t.Error("Unexpected Yield arguments result ", yield.(*Yield).Arguments, v)
	}
}

func TestCountSessions(t *testing.T) {
	registerSessions()
	invocation := &Invocation{
		Request: ID(1234),
	}
	yield, err := testRouter.countSessions(invocation)
	if err != nil {
		t.Error("Error on listSessions invocation ", err)
	}

	if yield.(*Yield).Arguments[0].(int) != 3 {
		t.Error("Unexpected Yield arguments result ", yield.(*Yield).Arguments)
	}
}

func TestGetSessions(t *testing.T) {
	registerSessions()
	invocation := &Invocation{
		Request:   ID(1234),
		Arguments: []interface{}{"fakeSession_0"},
	}
	yield, err := testRouter.getSession(invocation)
	if err != nil {
		t.Error("Error on listSessions invocation ", err)
	}

	v, ok := yield.(*Yield).ArgumentsKw["subscriptions"]
	if !ok {
		t.Error("Unexpected Yield subscriptions ArgumentsKW not found ", yield.(*Yield).ArgumentsKw)
	}
	if len(v.(map[string]interface{})) != 0 {
		t.Error("Unexpected Yield subscriptions lenght", yield.(*Yield).ArgumentsKw)
	}
	v, ok = yield.(*Yield).ArgumentsKw["registrations"]
	if !ok {
		t.Error("Unexpected Yield registrations ArgumentsKW not found ", yield.(*Yield).ArgumentsKw)
	}
	if len(v.(map[string]interface{})) != 0 {
		t.Error("Unexpected Yield registrations lenght ", yield.(*Yield).ArgumentsKw)
	}
}

func registerSessions() {
	testRouter = &DefaultRouter{
		sessions: make(map[PeerID]*Session),
		mutex:    &sync.RWMutex{},
	}

	// register 3 sessions
	for i := 0; i < 3; i++ {
		fp := NewFakePeer(PeerID(fmt.Sprintf("fakeSession_%d", i)))
		session := NewSession(fp)

		testRouter.register(session)
	}
}
