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
		Request:   ID(1234),
		Procedure: URI("foo"),
		Arguments: []interface{}{"bar", 1},
	}

	rsp := d.Call(call, fp)
	if rsp.MsgType() == ERROR {
		t.Error("Error executing Call ", rsp.(*Error).Error)
	}
	response := rsp.(*Result)
	if response.Request != ID(1234) {
		t.Error("Unexpected Result response")
	}
	if response.Arguments[0] != "okiDoki" {
		t.Error("Unexpected Arguments Result response")
	}
}

// A peer requests Register its own handler as a regular uri
// when someone calls this uri forwards call to uri peer owner,
// handle result and forward it to requester peer
func TestDealerDelegatedCallsConcept(t *testing.T) {
	// External Peer, Registers external URI
	reg := &Register{
		Request:   ID(123456789),
		Procedure: URI("externalFoo"),
	}
	d := NewDealer()
	ep := NewFakePeer(PeerID("6666666"))
	response := d.RegisterExternalHandler(reg, ep)
	if response.MsgType() != REGISTERED {
		t.Error("Unexpected response type ", response.MsgType())
	}

	// simulates to handle registered URI and return result
	go func(exp Peer) {
		for {
			select {
			case msg := <-exp.Receive():
				// External result forward using listners
				d.ExternalResult(msg, exp)

			case msg := <-exp.(*fakePeer).snd:
				rsp, err := externalHandler(msg, exp)
				if err != nil {
					t.Error("Error executing ", err)
				}
				exp.(*fakePeer).rcv <- rsp
			}
		}
	}(ep)

	call := &Call{
		Request:   ID(2222),
		Procedure: URI("externalFoo"),
	}

	//regular peer calls external URI
	p := NewFakePeer(PeerID("123"))
	rsp := d.Call(call, p)
	if rsp.MsgType() == ERROR {
		t.Error("Error executing Call ", rsp.(*Error).Error)
	}

	if rsp.MsgType() != RESULT {
		t.Error("Unexpected msg type ", rsp.MsgType())
	}

	if rsp.(*Result).Arguments[0] != "external" {
		t.Error("Unexpected result content")
	}
}

func fooHandler(msg Message, p Peer) (Message, error) {
	call := msg.(*Call)
	return &Result{
		Request:   call.Request,
		Arguments: []interface{}{"okiDoki"},
	}, nil
}

func externalHandler(msg Message, p Peer) (Message, error) {
	call := msg.(*Call)
	return &Result{
		Request:   call.Request,
		Arguments: []interface{}{"external"},
	}, nil
}
