package core

import (
	"testing"
	"time"
"log"
	"net/url"
"github.com/gorilla/websocket"
)

func TestServerConnectionHandling(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	s := NewServer(8888)
	go s.Run()

	time.Sleep(time.Millisecond * 100)
	tstClient := NewTestClient("localhost:8888")
	go tstClient.readLoop()
	time.Sleep(time.Millisecond * 100)
	tstClient.Send(&Hello{Realm: "foo", Details: map[string]interface{}{"foo": "bar"}})
	r := <- tstClient.rsp
	if r.MsgType() != WELCOME {
		t.Error("unexpected Hello response ", r.MsgType())
	}
	log.Println("Received from subscribe ", r)

	subs := &Subscribe{Request:ID(123), Topic: Topic("foo")}
	tstClient.Send(subs)
	r = <- tstClient.rsp
	if r.MsgType() != SUBSCRIBED {
		t.Error("unexpected Hello response ", r.MsgType())
	}
	log.Println("Received from subscribe ", r)

	s.Terminate()
	close(tstClient.done)
}

type testClient struct {
	Peer
	subscriptions map[ID]bool
	rsp chan Message
	done chan struct{}
}

func NewTestClient(host string) *testClient{
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	log.Printf("connected to %s \n", u.String())

	sc := &testClient{
		Peer: NewWebsockerPeer(conn, CLIENT),
		rsp: make(chan Message),
		done: make(chan struct{}),
	}

	go sc.readLoop()

	return sc
}

func (c *testClient)readLoop() {
	for {
		select {
		case msg, open := <- c.Receive():
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