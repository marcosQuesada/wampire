package wampire

import (
	"testing"
)

func TestBrokerPublish(t *testing.T) {
	b := NewBroker()
	fp := NewFakePeer(PeerID("123"))

	subs := &Subscribe{Request:ID(123), Topic: Topic("foo")}
	r := b.Subscribe(subs, fp)

	if r.MsgType() != SUBSCRIBED {
		t.Error("Error subscribing")
	}

	fp2 := NewFakePeer(PeerID("1234"))

	subs = &Subscribe{Request:ID(1234), Topic: Topic("foo")}
	b.Subscribe(subs, fp2)

	b.Publish(&Publish{Request:ID(9999), Topic:Topic("foo")}, fp)

	answ := <- fp2.snd
	if answ.MsgType() != PUBLISH {
		t.Error("Unexpected publish type ", answ.MsgType())
	}
}

func TestBrokerSubscribe(t *testing.T) {
	b := NewBroker()
	fp := NewFakePeer(PeerID("123"))

	subs := &Subscribe{Request:ID(123), Topic: Topic("foo")}
	r := b.Subscribe(subs, fp)

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

	if _, ok := b.topicBySubscription[subsRes.Subscription]; !ok {
		t.Error("Unexpected topic subscription not found")
	}
}

func TestBrokerUnSubscribe(t *testing.T) {
	b := NewBroker()
	fp := NewFakePeer(PeerID("123"))

	subs := &Subscribe{Request:ID(123), Topic: Topic("foo")}
	r := b.Subscribe(subs, fp)

	if r.MsgType() != SUBSCRIBED {
		t.Error("Error subscribing")
	}
	subsRes := r.(*Subscribed)
	unSubs := &Unsubscribe{Request:ID(123), Subscription: subsRes.Subscription}
	ru := b.UnSubscribe(unSubs, fp)
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

	if len(b.topicBySubscription) != 0 {
		t.Error("Unexpected subscriptions size")
	}
}