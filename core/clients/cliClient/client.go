package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/marcosQuesada/wampire/core"
	"github.com/olekukonko/tablewriter"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type clientMsgHandler func(core.Message) error
type cliClient struct {
	*core.Session

	msgHandlers map[core.MsgType]clientMsgHandler
	reader      *bufio.Reader
	done        chan struct{}
}

func NewCliClient(host string) *cliClient {
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	log.Printf("connected to %s \n", u.String())
	peer := core.NewWebsockerPeer(conn, core.CLIENT)

	c := &cliClient{
		Session: core.NewSession(peer),
		reader:  bufio.NewReader(os.Stdin),
		done:    make(chan struct{}),
	}

	// Declare local handlers
	c.msgHandlers = map[core.MsgType]clientMsgHandler{
		core.WELCOME: c.welcome,
		core.RESULT:  c.result,
		core.EVENT:   c.event,
	}

	realm := core.URI("fooRealm")
	details := map[string]interface{}{"foo": "bar"}
	c.sayHello(realm, details)
	go c.receiveLoop()

	return c
}

func (c *cliClient) RegisterMsgHandler(m core.MsgType, h clientMsgHandler) {
	c.msgHandlers[m] = h
}

func (c *cliClient) sayHello(realm core.URI, details map[string]interface{}) {
	c.Send(&core.Hello{Realm: realm, Details: details})
}

func (c *cliClient) processCli() {
	for {
		args, err := c.parseLine()
		if err != nil {
			log.Println("Error reading Cli line ", err)
			continue
		}
		msg := strings.ToUpper(args[0])
		switch msg {
		case "HELP":
			fmt.Fprint(os.Stdout, "Commands: \n")
			fmt.Fprint(os.Stdout, "  SUB Topic \n")
			fmt.Fprint(os.Stdout, "  PUB Topic \n")
			fmt.Fprint(os.Stdout, "  EXIT \n")
		case "PUB":
			if len(args) < 3 {
				log.Println("Invalid parameters: PUB Topic Message")
				continue
			}

			message := strings.Join(args[2:], " ")
			item := map[string]interface{}{"nick": "CliSession", "message": message}
			pub := &core.Publish{
				Request:   core.NewId(),
				Options:   map[string]interface{}{"acknowledge": true},
				Topic:     core.Topic(args[1]),
				Arguments: []interface{}{item},
			}
			c.Send(pub)
		case "SUB":
			if len(args) == 1 {
				log.Println("SUB Void Topic")
				continue
			}
			id := core.NewId()
			subs := &core.Subscribe{Request: id, Topic: core.Topic(args[1])}
			c.Send(subs)
		case "CALL":
			if len(args) == 1 {
				log.Println("CALL Void URI")
				continue
			}
			call := &core.Call{
				Request:   core.NewId(),
				Procedure: core.URI(args[1]),
				Arguments: []interface{}{"bar", 1},
			}
			c.Send(call)
		case "ID":
			log.Println("I am ", c.ID())
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

			if handler, ok := p.msgHandlers[msg.MsgType()]; ok {
				err := handler(msg)
				if err != nil {
					log.Println("Error executing client message handler")
				}
				continue
			}
			log.Println("Unhandled client message: ", msg.MsgType())

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

func (c *cliClient) printTableFromList(key string, values []interface{}) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{key, "value"})
	for _, v := range values {
		entry := []string{key, fmt.Sprintf("%s", v)}
		table.Append(entry)
	}
	table.Render()
}

// Expects map[string]map[string]interface{}
func (c *cliClient) printTableFromMap(m map[string]interface{}) {
	// iterate on each result set
	for mainKey, samples := range m {
		table := tablewriter.NewWriter(os.Stdout)
		if values, ok := samples.(map[string]interface{}); ok {
			for key, value := range values {
				entry := []string{fmt.Sprintf("%s", key), fmt.Sprintf("%s", value)}
				table.Append(entry)
			}
			if len(values) != 0 {
				table.SetHeader([]string{mainKey, "Value"})
				table.Render()
			}
		}
		// if entry samples ara a list handle it
		if mv, ok := samples.([]interface{}); ok {
			c.printTableFromList(mainKey, mv)
		}
	}
}

func (p *cliClient) welcome(msg core.Message) error {
	r := msg.(*core.Welcome)
	log.Printf("Welcome Details: %s \n", r.Details)
	return nil
}

func (p *cliClient) result(msg core.Message) error {
	r := msg.(*core.Result)
	log.Printf("RESULT %d", r.Request)
	if len(r.ArgumentsKw) > 0 {
		//format table results
		p.printTableFromMap(r.ArgumentsKw)
	}
	if len(r.Arguments) > 0 {
		p.printTableFromList("list", r.Arguments)
	}

	return nil
}

func (p *cliClient) event(msg core.Message) error {
	r := msg.(*core.Event)
	if len(r.Arguments) > 0 {
		var message string
		if detailsMap, ok := r.Arguments[0].(map[string]interface{}); ok {
			message = detailsMap["message"].(string)
		}
		log.Printf("EVENT Topic: %s Message: %s \n", r.Details["topic"], message)
	}

	return nil
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
