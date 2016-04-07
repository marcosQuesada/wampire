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

type SessionMetaEventHandler struct {
	metaEvents chan *MetaEvent
	done       chan struct{}
	mutex      *sync.Mutex
}

func NewSessionMetaEventsHandler() *SessionMetaEventHandler {
	return &SessionMetaEventHandler{
		metaEvents: make(chan *MetaEvent),
		done:       make(chan struct{}),
		mutex:      &sync.Mutex{},
	}
}

func (s *SessionMetaEventHandler) fireMetaEvents(id PeerID, message URI, details map[string]interface{}) {
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

func (s *SessionMetaEventHandler) consumeMetaEvents(r *Router) {
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
