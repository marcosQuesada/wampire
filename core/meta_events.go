package core

import (
	"log"
	"sync"
)

type MetaEvent struct {
	topic   Topic
	peerID  PeerID
	msg     URI
	details map[string]interface{}
}

type SessionMetaEventHandler interface {
	Fire(PeerID, URI, map[string]interface{})
	Consume(r *DefaultRouter)
	Terminate()
}

type defaultSessionMetaEventHandler struct {
	metaEvents chan *MetaEvent
	done       chan struct{}
	mutex      *sync.Mutex
}

func NewSessionMetaEventsHandler() *defaultSessionMetaEventHandler {
	return &defaultSessionMetaEventHandler{
		metaEvents: make(chan *MetaEvent),
		done:       make(chan struct{}),
		mutex:      &sync.Mutex{},
	}
}

func (s *defaultSessionMetaEventHandler) Fire(id PeerID, message URI, details map[string]interface{}) {
	//Fire on_join Session Meta Event only if is not the internal peer
	if id == PeerID("internal") {
		return
	}
	// fired in a non blocking way
	go func() {
		s.metaEvents <- &MetaEvent{
			topic:   Topic("wampire.session.meta.events"),
			peerID:  id,
			msg:     message,
			details: details,
		}
	}()
}

func (s *defaultSessionMetaEventHandler) Consume(r *DefaultRouter) {
	defer log.Println("Closed fireMetaEvents Loop")
	for {
		select {
		case mec, open := <-s.metaEvents:
			if !open {
				return
			}

			r.Broker.Publish(
				&Publish{
					Request: NewId(),
					Topic:   mec.topic,
					Options: map[string]interface{}{
						"session_id":  mec.peerID,
						"acknowledge": true,
						"details":     mec.details,
					},
					Arguments: []interface{}{
						map[string]interface{}{"message": mec.msg},
					},
				}, r.internalSession.session)
		case <-s.done:
			return

		}
	}
}

func (s *defaultSessionMetaEventHandler) Terminate() {
	close(s.done)
}

/** Fake Session Meta Events Handler to be used as Stub **/
type fakeSessionMetaEventsHandler struct{}

func (*fakeSessionMetaEventsHandler) Fire(PeerID, URI, map[string]interface{}) {}
func (*fakeSessionMetaEventsHandler) Consume(r *DefaultRouter)                 {}
func (*fakeSessionMetaEventsHandler) Terminate()                               {}
