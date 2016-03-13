package wampire

import (
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
)

type WebSocketClient struct {
	conn     *websocket.Conn
	snd      chan Message
	rcv      chan Message
	response chan Message
	done     chan struct{}
	exit     chan struct{}
	serializer Serializer
}

func NewWebSocketClient(host string) *WebSocketClient {
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	w := &WebSocketClient{
		conn:     c,
		snd:      make(chan Message),
		rcv:      make(chan Message),
		response: make(chan Message),
		done:     make(chan struct{}),
		exit:     make(chan struct{}),
		serializer: &JsonSerializer{},
	}

	go w.run()
	go w.readLoop()
	go w.pingLoop()

	return w
}

func (w *WebSocketClient) Send(m Message) {
	w.snd <- m
}

func (w *WebSocketClient) Receive() chan Message {
	return w.response
}

func (w *WebSocketClient) Terminate() {
	log.Println("Websocket Client terminate invoked")
	close(w.exit)
	close(w.done)
	close(w.response)
}

func (w *WebSocketClient) run() {
	defer log.Println("Exit Run loop")
	for {
		select {
		case msg, open := <-w.rcv:
			if !open {
				log.Println("Websocket Chann rcv closed, return ")
				return
			}
			log.Println("Websocket Chann rcv ", msg)
			//Must discard PONG messages
			w.response <- msg
		case msg, open := <-w.snd:
			if !open {
				log.Println("Websocket Chann send closed, return ")
				return
			}
			log.Println("Websocket Chann send ", msg)
			data, err := w.serializer.Serialize(msg)
			if err != nil {
				log.Fatal(err)
			}
			err = w.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Error writing userId as first message:", err)
				return
			}
		case <-w.done:
			log.Println("Exiting Websocket Run")
			return
		}
	}
}

func (w *WebSocketClient) readLoop() {
	defer w.Terminate()
	defer log.Print("Exiting Read Loop")

	for {
		_, data, err := w.conn.ReadMessage()
		if err != nil {
			log.Println("read loop Error:", err)
			return
		}
		message, err := w.serializer.Deserialize(data)
		if err != nil {
			log.Fatal("Fatal on deserialize ", err)
		}
		log.Printf("recv: %s", message)
		w.rcv <- message
	}
}

func (w *WebSocketClient) pingLoop() {
	pingTicker := time.NewTicker(time.Second * 5)
	defer log.Println("Exit Ping loop")
	defer w.conn.Close()
	defer pingTicker.Stop()

	for {
		select {
		case <-pingTicker.C:
			log.Println("Writting Ping")
			err := w.conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-w.exit:
			log.Println("PingLoop exit closed")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-w.done:
				log.Println("getting done closed on ping loop")
			case <-time.After(time.Second * 1):
				log.Println("getting timeout on exit ping loop")
			}
			return
		}
	}
}
