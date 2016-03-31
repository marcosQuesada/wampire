package core

import (
	"log"
	"testing"
	"time"
)

func TestDealerCallOnInternalPeer(t *testing.T) {
	m := &SessionMetaEventHandler{
		metaEvents: make(chan *MetaEvent, 10),
	}
	d := NewDealer(m)
	fp := NewFakePeer(PeerID("123"))
	s := NewSession(fp)
	go sessionLoop(s)

	i := NewInternalPeer()
	si := NewSession(i)

	// Consume internal session Receive channel
	go func(exp *Session) {
		for {
			select {
			case msg := <-exp.Receive():
				d.Yield(msg, s)
			}
		}
	}(si)

	uri := URI("foo")
	err := si.register(uri, fooHandler)
	if err != nil {
		t.Error("Unexpected error registering handler", err)
	}
	msg := d.Register(&Register{Request: ID(123), Procedure: uri}, si)
	if msg.MsgType() != REGISTERED {
		t.Error("Unexpected Register response ", msg.MsgType())
	}

	//give enough time to register handler
	time.Sleep(time.Millisecond * 200)

	// check uri for registration on session
	registeredUri, err := si.uriFromRegistration(msg.(*Registered).Registration)
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
	rsp := d.Call(call, s)
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
}

func sessionLoop(s *Session) {
	for {
		select {
		case msg := <-s.Receive():
			log.Println("client session receive ", msg.MsgType(), s.ID())
		}
	}
}

// A peer requests Register its own handler as a regular uri
// when someone calls this uri forwards call to uri peer owner,
// handle result and forward it to requester peer

func fooHandler(msg Message) (Message, error) {
	inv := msg.(*Invocation)
	return &Yield{
		Request:   inv.Request,
		Arguments: []interface{}{"okiDoki"},
	}, nil
}
