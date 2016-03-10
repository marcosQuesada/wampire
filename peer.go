package wamp

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	writeWait      = 1 * time.Second
	pongWait       = 10 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024
)

type Peer interface {
	Send(Message)
	Receive() chan Message
	ID() PeerID
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type webSocketPeer struct {
	serializer Serializer
	receive    chan Message
	send       chan Message
	exit       chan struct{}
	conn       *websocket.Conn
	id         PeerID
}

func NewWebsockerPeer(conn *websocket.Conn) *webSocketPeer {
	p := &webSocketPeer{
		serializer: &JsonSerializer{},
		receive:    make(chan Message),
		send:       make(chan Message),
		exit:       make(chan struct{}),
		conn:       conn,
		id:         PeerID(NewId()),
	}
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

func (p *webSocketPeer) ID() PeerID {
	return p.id
}

func (p *webSocketPeer) terinate() {
	close(p.exit)
}

func (p *webSocketPeer) writeLoop() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		p.conn.Close()
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
				return
			}
		case <-p.exit:
			break
		}
	}
}

func (p *webSocketPeer) readLoop() {
	defer func() {
		p.conn.Close()
	}()

	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPingHandler(func(string) error {
		log.Println("Received Ping, renewing deadline")
		p.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := p.conn.ReadMessage()
		if err != nil {
			break
		}

		log.Println("Received ", string(data))
		message, err := p.serializer.Deserialize(data)
		if err != nil {
			log.Fatal("Fatal on deserialize ", err)

		}
		if message != nil {
			p.receive <- message
		}
	}
}

func (p *webSocketPeer) write(mt int, message []byte) error {
	p.conn.SetWriteDeadline(time.Now().Add(writeWait))

	return p.conn.WriteMessage(mt, message)
}
