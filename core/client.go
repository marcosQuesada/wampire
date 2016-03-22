package core

import (
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

type Client struct {
	Peer
	subscriptions map[ID]bool
	msgHandlers   map[MsgType]Handler
	uriHandlers   map[URI]Handler
	exit          chan struct{}
}

func NewClient(host string) *Client {
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}
	log.Printf("connecting to %s \n", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	pc := &Client{
		Peer:          NewWebsockerPeer(conn),
		subscriptions: make(map[ID]bool),
		msgHandlers:   make(map[MsgType]Handler),
		uriHandlers:   make(map[URI]Handler),
		exit:          make(chan struct{}),
	}

	go pc.run()
	pc.sayHello()

	return pc
}

//Not requires concurrent access
func (p *Client) RegisterURI(u URI, h Handler) {
	p.uriHandlers[u] = h
}
func (p *Client) Register(m MsgType, h Handler) {
	p.msgHandlers[m] = h
}

func (p *Client) Exit() {
	close(p.exit)
	p.Terminate()
}

func (p *Client) sayHello() {
	p.Send(&Hello{Realm:"fooRealm", Details: map[string]interface{}{"foo":"bar"}}) //@TODO: Handle details
}

func (p *Client) run() {
	defer log.Println("Exit Run loop")

	for {
		select {
		case msg, open := <-p.Receive():
			if !open {
				log.Println("Websocket Chann rcv closed, return ")
				return
			}
			log.Println("Websocket Chann rcv ", msg)
			response := p.route(msg)
			if response != nil {
				p.Send(response)
			}

		case <-p.exit:
			return
		}
	}
}

func (p *Client) route(msg Message) Message {
	switch msg.MsgType() {
	case WELCOME:
		log.Println("Client received Welcome ", msg.(*Welcome).Id)
		return nil
	case ABORT:
		log.Println("Client received Abort")
		return nil
	case PUBLISH:
		log.Println("Received Publish ", msg)
		return nil
	case PUBLISHED:
		log.Println("Received Published ", msg)
		return nil
	case SUBSCRIBED:
		log.Println("Received Subscribed, subscription ID: ", msg.(*Subscribed).Subscription)
		p.subscriptions[msg.(*Subscribed).Subscription] = true
		return nil
	case UNSUBSCRIBED:
		log.Println("Received UnSubscribed, subscription Req ID: ", msg.(*Unsubscribed).Request)
		//delete(p.subscriptions, msg.(*Unsubscribed).)
		return nil
	case CALL:
		response := p.handleCall(msg.(*Call))
		log.Println("Executed call %s result ", msg.(*Call).Procedure, response)
		p.Send(response)
		return nil
	case ERROR:
		log.Println("Error", msg.(*Error).Error)
		close(p.exit)
		return nil
	}

	uri := "Message type not handled"
	log.Print(uri, msg.MsgType())

	return &Error{
		Error: URI(uri),
	}

}

func (p *Client) handleCall(call *Call) Message {
	handler, ok := p.uriHandlers[call.Procedure]
	if !ok {
		uri := "Client Handler not found"
		log.Print(uri)
		return &Error{
			Error: URI(uri),
		}
	}

	response, err := handler(call, p)
	if err != nil {
		uri := "Client Error invoking Handler"
		log.Print(uri, err)

		return &Error{
			Error: URI(uri),
		}
	}

	return response
}
