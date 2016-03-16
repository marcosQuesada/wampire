package wampire

import (
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

type PeerClient struct {
	Peer
	subscriptions map[ID]bool
	handlers      map[URI]Handler
	exit          chan struct{}
}

func NewPeerClient(host string) *PeerClient {
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	pc := &PeerClient{
		Peer:          NewWebsockerPeer(conn),
		subscriptions: make(map[ID]bool),
		handlers:      make(map[URI]Handler),
		exit:          make(chan struct{}),
	}

	go pc.run()
	pc.sayHello()

	return pc
}

//Not requires concurrent access
func (p *PeerClient) Register(u URI, h Handler) {
	p.handlers[u] = h
}

func (p *PeerClient) Exit() {
	close(p.exit)
	p.Terminate()
}

func (p *PeerClient) sayHello() {
	p.Send(&Hello{Id: NewId()})
}

func (p *PeerClient) run() {
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
			//@TODO: Must be handled server side! Unsubscribe before going down
/*			for subscriptionId, _ := range p.subscriptions {
				u := &Unsubscribe{Request: NewId(), Subscription: subscriptionId}
				p.Send(u)
			}
			log.Println("Exiting Websocket Run")*/
			return
		}
	}
}

func (p *PeerClient) route(msg Message) Message {
	switch msg.MsgType() {
	case WELCOME:
		log.Println("Client received Welcome")
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
		log.Println("Received Subscribed ", msg)
		p.subscriptions[msg.(*Subscribed).Subscription] = true
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

func (p *PeerClient) handleCall(call *Call) Message {
	handler, ok := p.handlers[call.Procedure]
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
