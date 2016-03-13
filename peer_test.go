package wampire

import (
	"testing"
	/*	"net"
	"github.com/gorilla/websocket"
		"net/url"*/)

func TestPeerOverPipeCons(t *testing.T) {
	/*	a, b := net.Pipe()
		u := &url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
		connA, _, _ := websocket.NewClient(a, u, nil, 4096, 4096)
		peerA := NewWebsockerPeer(connA)

		connB, _, _ := websocket.NewClient(b, u, nil, 4096, 4096)
		peerB := NewWebsockerPeer(connB)

	//	peerA.send <- &Hello{}
	*/ /*	rcvB := <- peerB.receive
		if rcvB.MsgType() != HELLO {
			t.Error("Unexpected type")
		}
		peerB.send <- &Hello{}
		rcvA := <- peerA.receive
		if rcvA.MsgType() != HELLO {
			t.Error("Unexpected type")
		}*/ /*

		peerA.terinate()
		peerB.terinate()*/
}

type fakePeer struct {
	rcv chan Message
	snd chan Message
	id  PeerID
}

func NewFakePeer(id PeerID) *fakePeer {
	return &fakePeer{
		rcv: make(chan Message, 10),
		snd: make(chan Message, 10),
		id:  id,
	}
}

func (p *fakePeer) Send(m Message) {
	p.snd <- m
}

func (p *fakePeer) Receive() chan Message {
	return p.rcv
}
func (p *fakePeer) Request(Message) Message{
	return &Result{

	}
}

func (p *fakePeer) ID() PeerID {
	return p.id
}
func (p *fakePeer) Terminate() {

}
