package core

type fakePeer struct {
	rcv chan Message
	snd chan Message
	id  PeerID
}

func NewFakePeer(id PeerID) *fakePeer {
	return &fakePeer{
		rcv: make(chan Message, 10),
		snd: make(chan Message, 10),
		id:  id,
	}
}

func (p *fakePeer) Send(m Message) {
	go func() {
		p.snd <- m
	}()
}

func (p *fakePeer) Receive() chan Message {
	//return p.rcv
	return p.snd
}

func (p *fakePeer) ID() PeerID {
	return p.id
}
func (p *fakePeer) Terminate() {

}
