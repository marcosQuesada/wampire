package wamp

import (
	"fmt"
	"log"
	"time"
)

type Router struct {
	peers  map[ID]Peer
	broker Broker
	dealer Dealer
}

func NewRouter() *Router {
	return &Router{
		peers:  make(map[ID]Peer),
		broker: NewBroker(),
		dealer: NewDealer(),
	}
}

func (r *Router) Accept(p Peer) error {
	timeout := time.NewTimer(time.Second * 1)
	select {
	case rcvMessage := <-p.Receive():
		timeout.Stop()
		h, ok := rcvMessage.(*Hello)
		if !ok {
			err := fmt.Sprintf("Unexpected type on Accept: %d", rcvMessage.MsgType())
			log.Println(err)

			return fmt.Errorf(err)
		}

		//@TODO: handle here authentication
		fmt.Println("Hello data ", h.Details)
		//answer welcome
		welcome := &Welcome{
			Id: h.Id,
		}
		p.Send(welcome)

		//if all goes fine
		go r.handleSession(p)

		return nil
	case <-timeout.C:
		return fmt.Errorf("Timeout error waiting Hello Message")
	}
}

func (r *Router) handleSession(p Peer) {
	for {
		select {
		case msg, open := <-p.Receive():
			if !open {
				break
			}

			log.Println("Received message on handle Session", msg)
		}
	}
}
