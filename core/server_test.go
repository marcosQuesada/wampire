package core

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"testing"
	//"time"
	"time"
)

func TestServerRunOnBoot(t *testing.T) {

}

func TestServerConnectionHandling(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	s := NewServer(8888)
	go s.Run()

	time.Sleep(time.Millisecond * 100)
	tstClientA := NewTestClient("localhost:8888")
	err := tstClientA.handshake()
	if err!=nil {
		t.Error("Unexpected error on testclient 1 handshake", err)
	}
	tstClientB := NewTestClient("localhost:8888")
	err = tstClientB.handshake()
	if err!=nil {
		t.Error("Unexpected error on testclient 1 handshake", err)
	}
/*	topic := Topic("foo")
	err = tstClientA.subscribeTopic(topic)
	if err!=nil {
		t.Error("Unexpected error on testclient 1 handshake", err)
	}*/
/*	err = tstClientB.subscribeTopic(topic)
	if err!=nil {
		t.Error("Unexpected error on testclient 1 handshake", err)
	}*/
	/*

	go tstClientA.session.Send(&Publish{Request:ID(9999), Topic:topic, Options:map[string]interface{}{"foo":"bar"}})
	r := <- tstClientA.rsp
	if r.MsgType() != PUBLISHED {
		t.Error("unexpected Hello response ", r.MsgType())
	}
	*/

	/*
	r = <- tstClientB.rsp
	if r.MsgType() != EVENT {
		t.Error("unexpected Hello response ", r.MsgType())
	}
	rDetails := r.(*Event).Details
	v, ok := rDetails["foo"]
	if  !ok {
		t.Error("Event Details not found")
	}
	if v!= "bar" {
		t.Error("Unmatched Event Details")
	}
	*/

	time.Sleep(time.Second * 1)
	tstClientA.session.Terminate()
	tstClientB.session.Terminate()
	s.Terminate()
	close(tstClientA.done)
	close(tstClientB.done)
}

type testClient struct {
	session       *Session
	subscriptions map[ID]bool
	rsp           chan Message
	done          chan struct{}
}

func NewTestClient(host string) *testClient {
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	log.Printf("connected to %s \n", u.String())
	peer := NewWebsockerPeer(conn, CLIENT)
	session := NewSession(peer)
	sc := &testClient{
		session: session,
		rsp:     make(chan Message),
		done:    make(chan struct{}),
	}

	go sc.readLoop()

	return sc
}

func (c *testClient) readLoop() {
	for {
		select {
		case msg, open := <-c.session.Receive():
			if !open {
				return
			}
			log.Print("Client Receive msg ", msg.MsgType())

			c.rsp <- msg
		case <-c.done:
			return
		}
	}
}

func (c *testClient) handshake() error {
	go c.session.Send(&Hello{Realm: "foo", Details: map[string]interface{}{"foo": "bar"}})
	r := <-c.rsp
	if r.MsgType() != WELCOME {
		return fmt.Errorf("unexpected Hello response ", r.MsgType())
	}

	return nil
}
func (c *testClient) subscribeTopic(topic Topic) error {
	go c.session.Send(&Subscribe{Request: ID(123), Topic: topic})
	r := <-c.rsp
	if r.MsgType() != SUBSCRIBED {
		return fmt.Errorf("unexpected Hello response ", r.MsgType())
	}
	return nil
}
