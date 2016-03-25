package core

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
	"time"
"github.com/gorilla/websocket"
)

type Server struct {
	port   int
	router *Router
}
var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	Subprotocols:    []string{"wamp.2.json"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewServer(port int) *Server {
	router := NewRouter()

	return &Server{
		port:   port,
		router: router,
	}
}

func (s *Server) Run() {
	defer log.Println("Start EXIT!!!")
	log.Println("Server Starting")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/ws", s.ServeWs)

	port := fmt.Sprintf(":%d", s.port)

	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Println("Server Error Listening ", err)
		return
	}
	log.Println("Before Serve")

	err = http.Serve(ln, router)
	if err != nil {
		log.Panic("Server Error Serving ", err)
		return
	}
}

func (s *Server) Terminate() {
	s.router.Terminate()
	time.Sleep(time.Second * 1)
	// Quick and dirty way to stop http.Serve!
	os.Exit(0)
}

func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	log.Println("Serve websocket connection")
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error upgrading", 403)
		return
	}

	p := NewWebsockerPeer(ws, SERVER)
	s.router.Accept(p)
}
