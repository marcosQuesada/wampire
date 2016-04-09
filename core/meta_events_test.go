package core

import (
	"sync"
	"testing"
)

func TestMetaEventsFireAndConsume(t *testing.T) {
	m := NewSessionMetaEventsHandler()
	m.Fire(PeerID("fake"), URI("fake_test"), map[string]interface{}{"foo": "bar"})

	router := &DefaultRouter{
		sessions:        make(map[PeerID]*Session),
		mutex:           &sync.RWMutex{},
		Broker:          NewBroker(m),
		exit:            make(chan struct{}),
		metaEvents:      m,
		internalSession: newInSession(),
	}

	s := NewSession(NewFakePeer(PeerID("123")))
	subs := &Subscribe{Request: ID(123), Topic: Topic("wampire.session.meta.events")}
	router.Subscribe(subs, s)
	r := <-s.Receive()

	if r.MsgType() != SUBSCRIBED {
		t.Error("Error subscribing")
	}

	go m.Consume(router)

	r = <-s.Receive()
	if r.MsgType() != EVENT {
		t.Error("Error subscribing")
	}

	if r.(*Event).Arguments[0].(map[string]interface{})["message"].(URI) != URI("fake_test") {
		t.Error("Unexpected details")
	}
	if r.(*Event).Details["details"].(map[string]interface{})["foo"] != "bar" {
		t.Error("Unexpected details")
	}
	if r.(*Event).Details["session_id"].(PeerID) != PeerID("fake") {
		t.Error("Unexpected details")
	}
}
