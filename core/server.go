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
	httpCientPath string
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
		httpCientPath: "",
	}
}

func (s *Server) Run() {
	defer log.Println("Start EXIT!!!")
	log.Println("Server Starting")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/ws", s.ServeWs)

	// add http client if required
	if s.httpCientPath != "" {
		httpDir := http.Dir(fmt.Sprintf("core/%s", s.httpCientPath))
		htmlClient := http.StripPrefix("/", http.FileServer(httpDir))
		router.PathPrefix("/").Handler(htmlClient)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Println("Server Error Listening ", err)
		return
	}

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

func (s *Server) SetHttpClient(path string) {
	s.httpCientPath = path
}