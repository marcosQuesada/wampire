package wamp

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Server struct {
}

type websocketServer struct {
	port   int
	router *Router
}

func NewServer(port int) *websocketServer {
	router := NewRouter()

	return &websocketServer{
		port:   port,
		router: router,
	}
}

func (s *websocketServer) Start() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/ws", s.serveWs)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func (s *websocketServer) Run() {
	for {
		/*		select {
				case p := <-s.register:
					s.peers[p.userId] = p
					log.Println("Registered user ", p.userId)
					break

				case p := <-s.unregister:
					_, ok := s.peers[p.userId]
					if ok {
						delete(s.peers, p.userId)
						close(p.send)
					}
					break

				case r := <-s.publish:
					p, ok := s.peers[r.userId]
					if !ok {
						log.Println("UserId Not found")
						break
					}

					p.send <- []byte(r.payload)
				}*/
	}

}

func (s *websocketServer) Terminate() {

}

func (s *websocketServer) serveWs(w http.ResponseWriter, r *http.Request) {
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

	p := NewWebsockerPeer(ws)

	s.router.Accept(p)
}
