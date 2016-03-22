package core

import (
	"testing"
	"time"
"log"
)

func TestServerConnectionHandling(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	s := NewServer(8888)
	go s.Run()

	time.Sleep(time.Millisecond * 100)
	tstClient := NewTestClient("localhost:8888")
	go tstClient.readLoop()
	time.Sleep(time.Millisecond * 100)
	/*

	tstClient.client.Send(&Hello{Id:NewId()})
	r := <- tstClient.rsp
	fmt.Println("Received from hello ", r)
	if r.MsgType() != WELCOME {
		t.Error("unexpected Hello response ", r.MsgType())
	}
	subs := &Subscribe{Request:ID("123"), Topic: Topic("foo")}
	tstClient.client.Send(subs)
	r = <- tstClient.rsp
	if r.MsgType() != SUBSCRIBED {
		t.Error("unexpected Hello response ", r.MsgType())
	}
	fmt.Println("Received from subscribe ", r)
	*/

	close(tstClient.done)
	//@TODO : Solve this test!
	s.Terminate()
}

type testClient struct {
	client  *Client
	rsp chan Message
	done chan struct{}
}

func NewTestClient(host string) *testClient{
	cl := NewClient(host)
	sc := &testClient{
		client: cl,
		rsp: make(chan Message),
		done: make(chan struct{}),
	}

	go sc.readLoop()

	return sc
}

func (c *testClient)readLoop() {
	for {
		select {
		case msg, open := <- c.client.Receive():
			if !open {
				return
			}
			log.Print("Client Receive msg", msg.MsgType())

			c.rsp <- msg
		case <-c.done:
			return
		}
	}
}