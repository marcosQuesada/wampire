package core

import (
	"log"
	"testing"
	"time"
)

func TestDealerCallBasicFlowOnInternalPeerInvocation(t *testing.T) {
	m := &fakeSessionMetaEventsHandler{}
	d := NewDealer(m)
	sessionA := NewSession(NewFakePeer(PeerID("123")))
	done := make(chan struct{})
	//go sessionLoop(sessionA, done)

	i := NewInternalPeer()
	si := NewSession(i)

	// Consume internal session Receive channel
	go func(exp *Session, exit chan struct{}) {
		for {
			select {
			case msg := <-exp.Receive():
				go d.Yield(msg, sessionA)
			case <-exit:
				return
			}
		}
	}(si, done)

	uri := URI("foo")
	err := si.register(uri, fooHandler)
	if err != nil {
		t.Error("Unexpected error registering handler", err)
	}

	d.Register(&Register{Request: ID(123), Procedure: uri}, si)

	//give enough time to register handler
	time.Sleep(time.Millisecond * 200)

	var registration ID
	for r, _ := range si.registrations {
		registration = r
	}
	// check uri for registration on session
	registeredUri, err := si.uriFromRegistration(registration)
	if err != nil {
		t.Error("Error looking up uriFromRegistration ", err)
	}

	if registeredUri != URI("foo") {
		t.Error("Unexpected registred URI")
	}

	// make Call
	call := &Call{
		Request:   ID(1234),
		Procedure: URI("foo"),
		Arguments: []interface{}{"bar", 1},
	}
	d.Call(call, sessionA)

	rsp := <-sessionA.Receive()
	if rsp.MsgType() == ERROR {
		t.Error("Error executing Call ", rsp.(*Error).Error)
		return
	}

	// check response
	response := rsp.(*Result)
	if response.Request != ID(1234) {
		t.Error("Unexpected Result response")
	}
	if response.Arguments[0] != "okiDoki" {
		t.Error("Unexpected Arguments Result response")
	}
	log.Println("Response ", response)

	close(done)
}

// A peer requests Register its own handler as a regular uri
// when someone calls this uri forwards call to uri peer owner,
// handle result and forward it to requester peer
func fooHandler(msg Message) (Message, error) {
	log.Println("Called")
	inv := msg.(*Invocation)
	return &Yield{
		Request:   inv.Request,
		Arguments: []interface{}{"okiDoki"},
	}, nil
}
