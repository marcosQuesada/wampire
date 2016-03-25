package main

import (
	"flag"
	"github.com/marcosQuesada/wampire/core"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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

	// enable html Client
	s.SetHttpClient("clients/htmlClient/")
	s.Run()
}
