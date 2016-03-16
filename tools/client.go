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

func main() {
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
	client *wampire.PeerClient
	done   chan struct{}
}

func NewTestClient(host string) *testClient {
	cl := wampire.NewPeerClient(host)
	sc := &testClient{
		client: cl,
		done:   make(chan struct{}),
	}


	fmt.Println("Subscribe to Topic foo")
	id := wampire.NewId()
	subs := &wampire.Subscribe{Request: id, Topic: wampire.Topic("foo")}
	sc.client.Send(subs)

	return sc
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
