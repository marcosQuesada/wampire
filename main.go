package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"github.com/marcosQuesada/wampire/core"
"github.com/gorilla/mux"
	"fmt"
	"net"
"net/http"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//Parse config
	port := flag.Int("port", 8888, "port")
	logOut := flag.Bool("log", false, "logger out path")
	flag.Parse()

	//Init logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *logOut {
		f, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Panic("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}
	s := core.NewServer(*port)
	c := make(chan os.Signal, 1)

	signal.Notify(
		c,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//serve until signal
	go func() {
		<-c
		s.Terminate()
	}()

	//Server Statics and Ws
	log.Println("Booting server on port ", *port)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/ws", s.ServeWs)
	htmlClient := http.StripPrefix("/", http.FileServer(http.Dir("clients/htmlClient/")))
	router.PathPrefix("/").Handler(htmlClient)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
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
