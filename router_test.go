package wampire

import (
	"testing"
)

func TestBasicRouterAccept(t *testing.T) {
	r := NewRouter()
	fp := NewFakePeer(PeerID("123"))
	go r.Accept(fp)

	fp.rcv <- &Hello{Id: ID("123")}

	w := <-fp.snd

	if w.MsgType() != WELCOME {
		t.Error("Unexpected welcome response")
	}

	if w.(*Welcome).Id != ID("123") {
		t.Error("Unexpected welcome ID response")
	}
}

func TestBasicRouterHandleSessionUnHandledMessage(t *testing.T) {
	r := NewRouter()
	fp := NewFakePeer(PeerID("123"))
	go r.handleSession(NewSession(fp))

	fp.rcv <- &Hello{Id: ID("123")}
	w := <-fp.snd

	if w.MsgType() != ERROR {
		t.Error("Unexpected welcome response")
	}
}
