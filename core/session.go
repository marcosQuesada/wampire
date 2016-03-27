package core

import (
	"fmt"
	"log"
	"sync"
)

type Handler func(Message) (Message, error)

type Session struct {
	Peer
	subscriptions map[ID]Topic    // Topic subscriptions
	registrations map[ID]URI      // Handler Registrations to URI
	handlers      map[URI]Handler // Handlers by URI
	mutex         *sync.RWMutex
}

func NewSession(p Peer) *Session {
	return &Session{
		Peer:          p,
		subscriptions: make(map[ID]Topic),
		registrations: make(map[ID]URI),
		handlers:      make(map[URI]Handler),
		mutex:         &sync.RWMutex{},
	}
}

// Goes to Internal peer
func (s *Session) register(uri URI, fn Handler) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.handlers[uri]; ok {
		return fmt.Errorf("%s handler already registered ", uri)
	}
	s.handlers[uri] = fn

	return nil
}

func (s *Session) unregister(uri URI) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.handlers[uri]; !ok {
		return fmt.Errorf("%s handler not registered ", uri)
	}
	delete(s.handlers, uri)

	return nil
}

func (s *Session) do(i *Invocation) error {
	log.Println("Doing session invocation ", i.Request, " registration ", i.Registration)
	uri, err := s.uriFromRegistration(i.Registration)
	if err != nil {
		log.Println("Error DO ", err, i.Registration)
		errUri := fmt.Sprintf("registration not found %s", uri)

		return fmt.Errorf(errUri)
	}

	// if handler is not register locally forward invocation to remote peer
	s.mutex.RLock()
	handler, ok := s.handlers[uri]
	s.mutex.RUnlock()
	if !ok {
		log.Println("Handler not found locally, forward it to remote peer ", s.ID(), "uri: ", uri)
		s.Send(i)

		return nil
	}

	// On local handling
	response, err := handler(i)
	if err != nil {
		return err
	}
	s.Send(response)

	return nil
}

func (s *Session) addRegistration(id ID, uri URI) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.registrations[id]; ok {
		return fmt.Errorf("%s registrations already registered on URI %s", id, uri)
	}
	s.registrations[id] = uri

	return nil
}

func (s *Session) removeRegistration(id ID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.registrations[id]; !ok {
		return fmt.Errorf("%s subscription not found", id)
	}
	delete(s.registrations, id)

	return nil
}

func (s *Session) uriFromRegistration(id ID) (URI, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	uri, ok := s.registrations[id]
	if !ok {
		return "", fmt.Errorf("%s registrations not found %s", id, uri)
	}

	return uri, nil
}

func (s *Session) addSubscription(id ID, topic Topic) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.subscriptions[id]; ok {
		return fmt.Errorf("%s subscription already registered on topic %s", id, topic)
	}
	s.subscriptions[id] = topic

	return nil
}

func (s *Session) removeSubscription(id ID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.subscriptions[id]; !ok {
		return fmt.Errorf("%s subscription not found", id)
	}
	delete(s.subscriptions, id)

	return nil
}
