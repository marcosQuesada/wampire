package core

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"testing"
	"time"
)

func TestServerConnectionHandlingAndShutDown(t *testing.T) {
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
	return &testClient{
		session: session,
		rsp:     make(chan Message),
		done:    make(chan struct{}),
	}
}

func (c *testClient) handshake() error {
	go c.session.Send(&Hello{Realm: "foo", Details: map[string]interface{}{"foo": "bar"}})
	r := <-c.session.Receive()
	if r.MsgType() != WELCOME {
		return fmt.Errorf("unexpected Hello response ", r.MsgType())
	}

	return nil
}
