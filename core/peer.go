package core

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	writeWait      = 1 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = (pongWait * 2) / 10
	maxMessageSize = 1024 * 1024
)

type Session struct {
	Peer
	subscriptions map[ID]Topic
}

func NewSession(p Peer) *Session {
	return &Session{
		Peer: p,
		subscriptions: make(map[ID]Topic),
	}
}

type Peer interface {
	Send(Message)
	Request(Message) Message
	Receive() chan Message
	ID() PeerID
	Terminate()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	Subprotocols:    []string{"wamp.2.json"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type webSocketPeer struct {
	id         PeerID
	conn       *websocket.Conn
	receive    chan Message
	send       chan Message
	closedConn chan struct{}
	exit       chan struct{}
	serializer Serializer
	wg         *sync.WaitGroup
}

func NewWebsockerPeer(conn *websocket.Conn) *webSocketPeer {
	p := &webSocketPeer{
		serializer: &JsonSerializer{},
		receive:    make(chan Message),
		send:       make(chan Message),
		exit:       make(chan struct{}),
		closedConn: make(chan struct{}),
		conn:       conn,
		id:         NewStringId(),
		wg:         &sync.WaitGroup{},
	}
	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPingHandler(func(string) error {
		log.Println("Received Ping, renewing deadline")
		p.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	p.wg.Add(2)
	go p.writeLoop()
	go p.readLoop()

	return p
}

func (p *webSocketPeer) Send(msg Message) {
	p.send <- msg
}

func (p *webSocketPeer) Receive() chan Message {
	return p.receive
}

func (p *webSocketPeer) Request(msg Message) Message {
	return &Error{}
}
func (p *webSocketPeer) ID() PeerID {
	return p.id
}

func (p *webSocketPeer) Terminate() {
	close(p.send)
	time.Sleep(time.Millisecond * 100) // give enough time to send close frame
	close(p.exit)

	p.conn.Close()
	p.wg.Wait()
	log.Println("webSocketPeer EXITED", string(p.id))
}

func (p *webSocketPeer) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.wg.Done()
	}()

	for {
		select {
		case message, ok := <-p.send:
			if !ok {
				p.write(websocket.CloseMessage, []byte{})
				return
			}
			data, err := p.serializer.Serialize(message)
			if err != nil {
				log.Fatal(err)
			}
			if err := p.write(websocket.TextMessage, data); err != nil {
				return
			}
		//ping message
		case <-ticker.C:
			if err := p.write(websocket.PingMessage, []byte{}); err != nil {
				log.Println("Error writting Ping message", err)
				return
			}
		//@TODO: Handle sync on one chan
		case <-p.closedConn:
			log.Println("writeLoop closedConn chan close")
			return
		case <-p.exit:
			log.Println("writeLoop exit chan close")
			return
		}
	}
}

func (p *webSocketPeer) readLoop() {
	defer func() {
		p.wg.Done()
		close(p.closedConn)
		close(p.receive)
	}()

	for {
		_, data, err := p.conn.ReadMessage()
		if err != nil {
			log.Println("Error reading Message on websocket Client", err)
			return
		}

		message, err := p.serializer.Deserialize(data)
		if err != nil {
			log.Fatal("Fatal on deserialize ", err)

		}
		p.receive <- message
	}
}

func (p *webSocketPeer) write(mt int, message []byte) error {
	p.conn.SetWriteDeadline(time.Now().Add(writeWait))

	return p.conn.WriteMessage(mt, message)
}
