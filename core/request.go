package core

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"
)

const TIMEOUT = time.Second * 1

type RequestListener struct {
	listeners map[ID]chan Message
	timeout   time.Duration
	mutex     sync.Mutex
}

func NewRequestListener() *RequestListener {
	return &RequestListener{
		listeners: make(map[ID]chan Message, 0),
		timeout:   TIMEOUT,
		mutex:     sync.Mutex{},
	}
}

func (r *RequestListener) Register(requestID ID) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.listeners[requestID] = make(chan Message, 1)
}

func (r *RequestListener) Notify(msg Message, requestID ID) {
	r.mutex.Lock()
	l, ok := r.listeners[requestID]
	r.mutex.Unlock()
	if ok {
		l <- msg
		return
	}
	log.Println("No listener found for request", requestID, "type", reflect.TypeOf(msg).String())
}

func (r *RequestListener) Wait(requestID ID) (msg Message, err error) {
	r.mutex.Lock()
	waitChannel, ok := r.listeners[requestID]
	r.mutex.Unlock()
	if !ok {
		return nil, fmt.Errorf("unknown listener ID: %v", requestID)
	} else {
		var timeout *time.Timer = time.NewTimer(time.Second * 1)
		select {
		case msg = <-waitChannel:
			timeout.Stop()
		case <-timeout.C:
			err = fmt.Errorf("timeout while waiting for message %s", requestID)
		}
	}
	close(waitChannel)
	delete(r.listeners, requestID)

	return
}

func (r *RequestListener) RegisterAndWait(requestID ID) (msg Message, err error) {
	r.Register(requestID)

	return r.Wait(requestID)
}
