package main

import (
	"fmt"
	"github.com/marcosQuesada/wampire"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"flag"
)

func mainClient() {
	//Init logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	//Parse config
	host := flag.String("Hostname", "localhost", "host name")
	port := flag.Int("port", 8888, "port")
	flag.Parse()

	c := make(chan os.Signal, 1)

	signal.Notify(
		c,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	tstClient := NewTestClient(fmt.Sprintf("%s:%d",*host, *port))

	//serve until signal
	go func() {
		<-c
		close(tstClient.done)
	}()

	tstClient.writeLoop()
}

type testClient struct {
	client *wampire.WebSocketClient
	done   chan struct{}
}

func NewTestClient(host string) *testClient {
	cl := wampire.NewWebSocketClient(host)
	sc := &testClient{
		client: cl,
		done:   make(chan struct{}),
	}

	go sc.run()

	return sc
}

func (c *testClient) run() {
	defer log.Println("Exiting client readLoop")

	id := wampire.NewId()
	c.client.Send(&wampire.Hello{Id: id})
	r := <-c.client.Receive()
	if r.MsgType() != wampire.WELCOME {
		log.Panic("unexpected Hello response ", r.MsgType())
	}
	fmt.Println("Subscribe to Topic foo")
	subs := &wampire.Subscribe{Request: id, Topic: wampire.Topic("foo")}
	c.client.Send(subs)
	r = <-c.client.Receive()
	if r.MsgType() != wampire.SUBSCRIBED {
		log.Panic("unexpected Hello response ", r.MsgType())
	}
	fmt.Println("Received from subscribe ", r)

	for {
		select {
		case msg, open := <-c.client.Receive():
			if !open {
				log.Println("Closed Rcv on readLoop")
				close(c.done)
				return
			}
			log.Print("Client Receive msg: ", msg.MsgType())

		case <-c.done:
			log.Println("Closed Done from readLoop")
			c.client.Terminate()
			return
		}
	}
}

func (c *testClient) writeLoop() {
	defer log.Println("Exiting client writeLoop")
	tick := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-tick.C:
			pub := &wampire.Publish{Request: wampire.NewId(), Topic: wampire.Topic("foo")}
			c.client.Send(pub)

		case <-c.done:
			log.Println("Closed Done from writeLoop")
			return
		}
	}
}
