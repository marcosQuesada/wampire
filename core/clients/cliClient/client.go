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
	"strconv"
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
		core.WELCOME:    c.welcome,
		core.PUBLISHED:  c.published,
		core.SUBSCRIBED: c.subscribed,
		core.RESULT:     c.result,
		core.INTERRUPT:  c.interrupt,
		core.YIELD:      c.yield,
		core.EVENT:      c.event,
		core.ERROR:      c.error,
	}

	// @TODO: Handle Hello details
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
			subs := &core.Subscribe{Request: id, Options: map[string]interface{}{}, Topic: core.Topic(args[1])}
			c.Send(subs)
		case "CALL":
			if len(args) == 1 {
				log.Println("CALL Void URI")
				continue
			}
			cappedArgs := args[2:]
			var newArgs []interface{}
			for i := 0; i < len(cappedArgs); i++ {
				newArgs = append(newArgs, cappedArgs[i])
			}
			call := &core.Call{
				Request:   core.NewId(),
				Procedure: core.URI(args[1]),
				Arguments: newArgs,
				Options:   map[string]interface{}{"receive_progress": true},
			}
			log.Println("Call request ", call.Request)
			c.Send(call)
		case "CANCEL":
			if len(args) != 2 {
				log.Println("CANCEL Void URI")
				continue
			}
			id, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				log.Println("Error parsing Request ID")
			}
			log.Println("Canceling task ", id)
			cancel := &core.Cancel{
				Request: core.ID(id),
				Options: map[string]interface{}{},
			}
			c.Send(cancel)
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
		entry := []string{key, fmt.Sprintf("%d", v)}
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
				stringValue := fmt.Sprintf("%s", value)
				entry := []string{fmt.Sprintf("%s", key), stringValue}
				table.Append(entry)
			}
			if len(values) != 0 {
				table.SetHeader([]string{mainKey, "Value"})
				table.Render()
			}
			continue
		}
		// if entry samples ara a list handle it
		if mv, ok := samples.([]interface{}); ok {
			c.printTableFromList(mainKey, mv)
			continue
		}

		fmt.Printf("%s: %s\n", mainKey, samples)
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

func (p *cliClient) interrupt(msg core.Message) error {
	r := msg.(*core.Interrupt)
	log.Printf("Interrupted Call Id: %s \n", r.Request)
	return nil
}

func (p *cliClient) yield(msg core.Message) error {
	r := msg.(*core.Yield)

	log.Printf("Yield Call Id: %d Update: %d  \n", r.Request, int(r.ArgumentsKw["update"].(float64)))
	return nil
}

func (p *cliClient) event(msg core.Message) error {
	r := msg.(*core.Event)
	if len(r.Arguments) > 0 {
		var message string
		if detailsMap, ok := r.Arguments[0].(map[string]interface{}); ok {
			message = detailsMap["message"].(string)
		}
		if cid, ok := r.Details["session_id"]; ok {
			message = fmt.Sprintf("%s id: %s", message, cid)
		}
		log.Printf("EVENT Topic: %s Message: %s \n", r.Details["topic"], message)
	}

	return nil
}

func (p *cliClient) published(msg core.Message) error {
	r := msg.(*core.Published)
	log.Printf("published Details: %d \n", r.Request)
	return nil
}

func (p *cliClient) subscribed(msg core.Message) error {
	r := msg.(*core.Subscribed)
	log.Printf("Subscribed Details: %d \n", r.Subscription)
	return nil
}

func (c *cliClient) exit() {
	close(c.done)
	c.Terminate()
	os.Exit(0)
}

func (p *cliClient) error(msg core.Message) error {
	r := msg.(*core.Error)
	log.Printf("Error URI: %s \n", r.Error)
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
		client.exit()
	}()

	client.processCli()
}
