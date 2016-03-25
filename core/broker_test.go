package core

import (
	"testing"
)

func TestBrokerPublish(t *testing.T) {
	b := NewBroker()
	fpp := NewFakePeer(PeerID("123"))
	s := NewSession(fpp)

	subs := &Subscribe{Request:ID(123), Topic: Topic("foo")}
	r := b.Subscribe(subs, s)

	if r.MsgType() != SUBSCRIBED {
		t.Error("Error subscribing")
	}

	fpp2 := NewFakePeer(PeerID("1234"))
	fp2 := NewSession(fpp2)

	subs = &Subscribe{Request:ID(1234), Topic: Topic("foo")}
	b.Subscribe(subs, fp2)

	b.Publish(&Publish{Request:ID(9999), Topic:Topic("foo")}, s)

	answ := <- fpp2.snd
	if answ.MsgType() != EVENT {
		t.Error("Unexpected publish type ", answ.MsgType())
	}
}

func TestBrokerSubscribe(t *testing.T) {
	b := NewBroker()
	s := NewSession(NewFakePeer(PeerID("123")))

	subs := &Subscribe{Request:ID(123), Topic: Topic("foo")}
	r := b.Subscribe(subs, s)

	if r.MsgType() != SUBSCRIBED {
		t.Error("Error subscribing")
	}

	subsRes := r.(*Subscribed)
	if subsRes.Request != ID(123) {
		t.Error("Unexpected response ID")
	}

	if len(b.topicPeers) != 1 {
		t.Error("Unexpected topicPeers size")
	}

	if len(b.topicPeers) != 1 {
		t.Error("Unexpected topicPeers size")
	}

	if len(b.topicPeers[Topic("foo")]) != 1 {
		t.Error("Unexpected topicPeers on peers size")
	}

	if subsRes.Subscription == 0 {
		t.Error(" Void SUbscription ID")
	}

	// assert Session subscription
	topic, ok := s.subscriptions[subsRes.Subscription]
	if  !ok {
		t.Error("Session topic subscription not found")
	}

	if topic != Topic("foo") {
		t.Error("Unexpected topic on session subscription")
	}

}

func TestBrokerUnSubscribe(t *testing.T) {
	b := NewBroker()
	s := NewSession(NewFakePeer(PeerID("123")))

	subs := &Subscribe{Request:ID(123), Topic: Topic("foo")}
	r := b.Subscribe(subs, s)

	if r.MsgType() != SUBSCRIBED {
		t.Error("Error subscribing")
	}
	subsRes := r.(*Subscribed)
	unSubs := &Unsubscribe{Request:ID(123), Subscription: subsRes.Subscription}
	ru := b.UnSubscribe(unSubs, s)
	if ru.MsgType() != UNSUBSCRIBED {
		t.Error("Error unsubscribing", ru.MsgType())
	}

	unSubsRes := ru.(*Unsubscribed)
	if unSubsRes.Request != ID(123) {
		t.Error("Unexpected response ID")
	}

	if len(b.topicPeers) != 0 {
		t.Error("Unexpected topicPeers size")
	}

	if len(b.subscriptions) != 0 {
		t.Error("Unexpected subscriptions size")
	}

	// assert Session unSubscription
	_, ok := s.subscriptions[subsRes.Subscription]
	if  ok {
		t.Error("Session topic subscription not found")
	}
}