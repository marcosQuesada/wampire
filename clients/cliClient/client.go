package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/marcosQuesada/wampire/core"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type cliClient struct {
	core.Peer
	subscriptions map[core.ID]bool
	msgHandlers   map[core.MsgType]core.Handler
	uriHandlers   map[core.URI]core.Handler
	reader        *bufio.Reader
	done          chan struct{}
}

func NewCliClient(host string) *cliClient {
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	log.Printf("connected to %s \n", u.String())

	pc := &cliClient{
		Peer:          core.NewWebsockerPeer(conn, core.CLIENT),
		subscriptions: make(map[core.ID]bool),
		msgHandlers:   make(map[core.MsgType]core.Handler),
		uriHandlers:   make(map[core.URI]core.Handler),
		reader:        bufio.NewReader(os.Stdin),
		done:          make(chan struct{}),
	}

	realm := core.URI("fooRealm")
	details := map[string]interface{}{"foo": "bar"}
	pc.sayHello(realm, details)
	go pc.receiveLoop()

	return pc
}

func (p *cliClient) Register(u core.URI, h core.Handler) {
	// concurrent access is not required
	p.uriHandlers[u] = h
}
func (p *cliClient) RegisterMsgHandler(m core.MsgType, h core.Handler) {
	p.msgHandlers[m] = h
}

func (p *cliClient) sayHello(realm core.URI, details map[string]interface{}) {
	p.Send(&core.Hello{Realm: realm, Details: details})
}

func (c *cliClient) processCli() {
	fmt.Print("Enter text: \n")
	for {
		fmt.Println(">")
		args, err := c.parseLine()
		if err != nil {
			log.Println("Error reading Cli line ", err)
			continue
		}
		msg := strings.ToUpper(args[0])
		log.Println("msg ", msg)
		switch msg {
		case "HELP":
			fmt.Fprint(os.Stdout, "Commands: \n")
			fmt.Fprint(os.Stdout, "  SUB Topic \n")
			fmt.Fprint(os.Stdout, "  PUB Topic \n")
			fmt.Fprint(os.Stdout, "  EXIT \n")
		case "PUB":
			if len(args) == 1 {
				log.Println("PUB Void Topic")
				continue
			}
			pub := &core.Publish{Request: core.NewId(), Options: map[string]interface{}{"foo": "bar"}, Topic: core.Topic(args[1])}
			c.Send(pub)
		case "SUB":
			if len(args) == 1 {
				log.Println("SUB Void Topic")
				continue
			}
			id := core.NewId()
			subs := &core.Subscribe{Request: id, Topic: core.Topic(args[1])}
			c.Send(subs)
		case "EXIT":
			log.Println("Exit Cli client")
			close(c.done)
			c.Terminate()
			return
		default:
			log.Println("Not handled Command", args)
		}
	}
}

func (p *cliClient) receiveLoop() {
	defer log.Println("Exit Run loop")

	for {
		select {
		case msg, open := <-p.Receive():
			if !open {
				log.Println("Websocket Chann rcv closed, return ")
				return
			}
			log.Println("Websocket Chann rcv ", msg.MsgType())
			log.Println(msg)
		case <-p.done:
			return
		}
	}
}

func (c *cliClient) parseLine() (args []string, err error) {
	rawLine, err := c.reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading Cli line ", err)
		return
	}
	args = strings.Split(rawLine, "\n")
	args = strings.Split(args[0], " ")

	if args[0] == "" {
		err = fmt.Errorf("Void message")
		return
	}

	return
}
func (p *cliClient) handleCall(call *core.Call) core.Message {
	handler, ok := p.uriHandlers[call.Procedure]
	if !ok {
		uri := "Client Handler not found"
		log.Print(uri)
		return &core.Error{
			Error: core.URI(uri),
		}
	}

	response, err := handler(call)//@TODO: Peer removed!
	if err != nil {
		uri := "Client Error invoking Handler"
		log.Print(uri, err)

		return &core.Error{
			Error: core.URI(uri),
		}
	}

	return response
}

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

	client := NewCliClient(fmt.Sprintf("%s:%d", *host, *port))

	//serve until signal
	go func() {
		<-c
		//@TODO....
		close(client.done)
		client.Terminate()
		os.Exit(0)
	}()

	client.processCli()
}
