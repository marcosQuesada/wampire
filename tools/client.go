package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/marcosQuesada/wampire"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

	client := NewCliClient(fmt.Sprintf("%s:%d", *host, *port))

	//serve until signal
	go func() {
		<-c
		close(client.done)
	}()

	client.processCli()
}

type cliClient struct {
	client *wampire.PeerClient
	reader *bufio.Reader
	done   chan struct{}
}

func NewCliClient(host string) *cliClient {
	cl := wampire.NewPeerClient(host)
	sc := &cliClient{
		client: cl,
		reader: bufio.NewReader(os.Stdin),
		done:   make(chan struct{}),
	}

	return sc
}

func (c *cliClient) processCli() {
	fmt.Print("Enter text: ")
	for {
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
			pub := &wampire.Publish{Request: wampire.NewId(), Options:map[string]interface{}{"foo":"bar"}, Topic: wampire.Topic(args[1] )}
			c.client.Send(pub)
		case "SUB":
			if len(args) == 1 {
				log.Println("SUB Void Topic")
				continue
			}
			id := wampire.NewId()
			subs := &wampire.Subscribe{Request: id, Topic: wampire.Topic(args[1])}
			c.client.Send(subs)
		case "EXIT":
			log.Println("Exit Cli client")
			c.client.Terminate()
			return
		default:
			log.Println("Not handled Command", args)
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
