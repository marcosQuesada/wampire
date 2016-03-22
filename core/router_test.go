package core

import (
	"testing"
)

func TestBasicRouterAccept(t *testing.T) {
	r := NewRouter()
	fp := NewFakePeer(PeerID("123"))
	go r.Accept(fp)

	details:= map[string]interface{}{
		"roles" : map[string]interface{}{"publisher":"", "subscriber":""},
	}
	fp.rcv <- &Hello{Details: details} //@TODO: assert Details

	w := <-fp.snd

	if w.MsgType() != WELCOME {
		t.Error("Unexpected welcome response")
	}

	if w.(*Welcome).Details == nil {
		t.Error("Unexpected welcome ID details")
	}
	//@TODO: Solve it!
/*	d, ok := w.(*Welcome).Details.(map[string]interface{})
	if !ok {
		t.Error("Unexpected details conversion")
	}
	if len(d["roles"]) != 1 {
		t.Error("Unexpected welcome ID details")
	}*/
}

func TestBasicRouterHandleSessionUnHandledMessage(t *testing.T) {
	r := NewRouter()
	fp := NewFakePeer(PeerID("123"))
	go r.handleSession(NewSession(fp))

	fp.rcv <- &Hello{Realm:"fooRealm", Details:map[string]interface{}{"foo":"bar"}}
	w := <-fp.snd

	if w.MsgType() != ERROR {
		t.Error("Unexpected welcome response")
	}
}
