package wampire

import (
	"testing"
)

func TestDealerCall(t *testing.T) {
	d := NewDealer()
	fp := NewFakePeer(PeerID("123"))
	err := d.Register(URI("foo"), fooHandler)
	if err != nil {
		t.Error("Unexpected error registering handler", err)
	}
	call := &Call{
		Request: ID("1234"),
		Procedure: URI("foo"),
		Arguments: []interface{}{"bar", 1},
	}

	rsp:= d.Call(call, fp)
	if rsp.MsgType() == ERROR {
		t.Error("Error executing Call ", rsp.(*Error).Error)
	}
	response := rsp.(*Result)
	if response.Request != ID("1234") {
		t.Error("Unexpected Result response")
	}
	if response.Arguments[0] != "okiDoki" {
		t.Error("Unexpected Arguments Result response")
	}
}

func fooHandler(msg Message, p Peer) (Message, error) {
	call := msg.(*Call)
	return &Result{
		Request: call.Request,
		Arguments: []interface{}{"okiDoki"},
	}, nil
}