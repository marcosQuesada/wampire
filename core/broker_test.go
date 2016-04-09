package core

import (
	"testing"
)

func TestBrokerPublish(t *testing.T) {
	// Mocked MetaEventsChannel
	m := &fakeSessionMetaEventsHandler{}
	b := NewBroker(m)

	sessionA := NewSession(NewFakePeer(PeerID("PeerA")))
	sessionB := NewSession(NewFakePeer(PeerID("PeerB")))

	// Subscribe Session A on Topic foo
	subs := &Subscribe{Request: ID(123), Topic: Topic("foo")}
	b.Subscribe(subs, sessionA)
	r := <-sessionA.Receive()
	if r.MsgType() != SUBSCRIBED {
		t.Error("Error subscribing ", r.MsgType())
	}

	// Subscribe Session B on Topic foo
	subs = &Subscribe{Request: ID(1234), Topic: Topic("foo")}
	b.Subscribe(subs, sessionB)
	r = <-sessionB.Receive()
	if r.MsgType() != SUBSCRIBED {
		t.Error("Error subscribing ", r.MsgType())
	}

	// Session A Publish on Topic foo
	b.Publish(&Publish{Request: ID(9999), Topic: Topic("foo")}, sessionA)
	r = <-sessionB.Receive()
	if r.MsgType() != EVENT {
		t.Error("Unexpected publish type ", r.MsgType())
	}
}

func TestBrokerSubscribe(t *testing.T) {
	m := &fakeSessionMetaEventsHandler{}
	b := NewBroker(m)
	s := NewSession(NewFakePeer(PeerID("123")))

	subs := &Subscribe{Request: ID(123), Topic: Topic("foo")}
	b.Subscribe(subs, s)
	r := <-s.Receive()
	if r.MsgType() != SUBSCRIBED {
		t.Error("Error subscribing")
	}

	subsRes := r.(*Subscribed)
	if subsRes.Request != ID(123) {
		t.Error("Unexpected response ID")
	}

	if len(b.topicPeers) != 2 {
		t.Error("Unexpected topicPeers size")
	}

	if len(b.topicPeers[Topic("wampire.session.meta.events")]) != 0 {
		t.Error("Unexpected topicPeers on peers size")
	}

	if len(b.topicPeers[Topic("foo")]) != 1 {
		t.Error("Unexpected topicPeers on peers size")
	}

	if subsRes.Subscription == 0 {
		t.Error(" Void SUbscription ID")
	}

	// assert Session subscription
	topic, ok := s.subscriptions[subsRes.Subscription]
	if !ok {
		t.Error("Session topic subscription not found")
	}

	if topic != Topic("foo") {
		t.Error("Unexpected topic on session subscription")
	}

}

func TestBrokerUnSubscribe(t *testing.T) {
	m := &fakeSessionMetaEventsHandler{}
	b := NewBroker(m)
	s := NewSession(NewFakePeer(PeerID("123")))

	subs := &Subscribe{Request: ID(123), Topic: Topic("foo")}
	b.Subscribe(subs, s)
	r := <-s.Receive()

	if r.MsgType() != SUBSCRIBED {
		t.Error("Error subscribing")
	}
	subsRes := r.(*Subscribed)
	unSubs := &Unsubscribe{Request: ID(123), Subscription: subsRes.Subscription}
	b.UnSubscribe(unSubs, s)
	ru := <-s.Receive()
	if ru.MsgType() != UNSUBSCRIBED {
		t.Error("Error unsubscribing", ru.MsgType())
	}

	unSubsRes := ru.(*Unsubscribed)
	if unSubsRes.Request != ID(123) {
		t.Error("Unexpected response ID")
	}

	if len(b.topicPeers[Topic("wampire.session.meta.events")]) != 0 {
		t.Error("Unexpected topicPeers on peers size")
	}

	// topic wampire.session.meta.events still exists
	if len(b.topicPeers) != 1 {
		t.Error("Unexpected topicPeers size")
	}

	if len(b.subscriptions) != 0 {
		t.Error("Unexpected subscriptions size")
	}

	// assert Session unSubscription
	_, ok := s.subscriptions[subsRes.Subscription]
	if ok {
		t.Error("Session topic subscription not found")
	}
}
