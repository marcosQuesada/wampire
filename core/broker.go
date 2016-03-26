package core

import (
	"fmt"
	"log"
	"sync"
)

type Broker interface {
	Subscribe(Message, *Session) Message
	UnSubscribe(Message, *Session) Message
	Publish(Message, *Session) Message
	Handlers() map[URI]Handler
}

type defaultBroker struct {
	topics        map[Topic]map[ID]bool   //maps topics to subscriptions
	subscriptions map[ID]*Session         //a peer may have many subscriptions
	topicPeers    map[Topic]map[PeerID]ID //maps peers by topic on subscription
	mutex         *sync.RWMutex
}

func NewBroker() *defaultBroker {
	return &defaultBroker{
		topics:        make(map[Topic]map[ID]bool),
		subscriptions: make(map[ID]*Session),
		topicPeers:    make(map[Topic]map[PeerID]ID),
		mutex:         &sync.RWMutex{},
	}
}

func (b *defaultBroker) Subscribe(msg Message, s *Session) Message {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	subscribe, ok := msg.(*Subscribe)
	if !ok {
		log.Fatal("Unexpected type on publish ", msg.MsgType())
		panic("Unexpected type on publish")
	}

	_, ok = b.topics[subscribe.Topic]
	if !ok {
		//create topic!
		b.topics[subscribe.Topic] = make(map[ID]bool)
		b.topicPeers[subscribe.Topic] = make(map[PeerID]ID)
	}

	//check if subscriptor is already register to topic
	if subs, ok := b.topicPeers[subscribe.Topic][s.ID()]; ok {
		log.Println("Session already subscribed on subscription ", subs)
		return &Error{
			Error: URI("Peer already subscribed on subscription"),
		}
	}

	subscriptionId := NewId()
	b.topicPeers[subscribe.Topic][s.ID()] = subscriptionId
	b.topics[subscribe.Topic][subscriptionId] = true
	b.subscriptions[subscriptionId] = s

	// Add subscription to session
	s.addSubscription(subscriptionId, subscribe.Topic)

	return &Subscribed{
		Request:      subscribe.Request,
		Subscription: subscriptionId,
	}
}

func (b *defaultBroker) UnSubscribe(msg Message, s *Session) Message {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	unsubscribe, ok := msg.(*Unsubscribe)
	if !ok {
		log.Fatal("Unexpected type on UnSubscribe ", msg.MsgType())
		panic("Unexpected type on UnSubscribe")
	}

	topic, ok := s.subscriptions[unsubscribe.Subscription]
	if !ok {
		uri := "topic not found to this Subscription"
		log.Println(uri, unsubscribe)

		return &Error{
			Request: unsubscribe.Request,
			Error:   URI(uri),
		}
	}

	session, ok := b.subscriptions[unsubscribe.Subscription]
	if !ok {
		uri := "peer not found to this Subscription"
		log.Println(uri, unsubscribe)

		return &Error{
			Request: unsubscribe.Request,
			Error:   URI(uri),
		}
	}

	//Remove session subscription
	s.removeSubscription(unsubscribe.Subscription)
	//remove peer from subscription map
	delete(b.subscriptions, unsubscribe.Subscription)
	//remove peer from topic map
	delete(b.topicPeers[topic], session.ID())
	//remove subscription from topic
	delete(b.topics[topic], unsubscribe.Subscription)

	//if void topic remove it
	if len(b.topics[topic]) == 0 {
		delete(b.topics, topic)
	}
	//if void topic remove it
	if len(b.topicPeers[topic]) == 0 {
		delete(b.topicPeers, topic)
	}

	return &Unsubscribed{
		Request: unsubscribe.Request,
	}
}

func (b *defaultBroker) Publish(msg Message, p *Session) Message {
	publish, ok := msg.(*Publish)
	if !ok {
		log.Fatal("Unexpected type on publish ", msg.MsgType())
	}

	b.mutex.RLock()
	subscribers, ok := b.topics[publish.Topic]
	b.mutex.RUnlock()
	if !ok {
		uri := "Topic not found"
		log.Println(uri, publish)

		return &Error{
			Request: publish.Request,
			Error:   URI(uri),
		}
	}

	//iterate on topic subscribers
	for subscriptionId, _ := range subscribers {
		b.mutex.RLock()
		//find peer from subscriber
		peer, ok := b.subscriptions[subscriptionId]
		b.mutex.RUnlock()
		if !ok {
			log.Println("Peer not found")
			continue
		}

		if peer.ID() != p.ID() {
			//  Publish to all topic subscriptors as event
			if publish.Options == nil {
				publish.Options = map[string]interface{}{}
			}
			publish.Options["topic"] = publish.Topic
			event := &Event{
				Subscription: subscriptionId,
				Publication:  publish.Request,
				Details:      publish.Options,
				Arguments:    publish.Arguments,
				ArgumentsKw:  publish.ArgumentsKw,
			}
			//send message in a non blocking way
			go peer.Send(event)
		}
	}

	return &Published{
		Request: publish.Request,
	}
}

func (b *defaultBroker) Handlers() map[URI]Handler {
	return map[URI]Handler{
		"wampire.core.broker.dump": b.dumpBroker,
	}
}

func (b *defaultBroker) dumpBroker(msg Message) (Message, error) {
	b.mutex.RLock()
	topics := b.topics
	subscriptions := b.subscriptions
	b.mutex.RUnlock()

	list := []interface{}{}
	for topic, _ := range topics {
		list = append(list, topic)
	}

	subs := map[string]interface{}{}
	for id, s := range subscriptions {
		subs[fmt.Sprintf("%d", id)] = s.ID()
	}
	inv := msg.(*Invocation)
	kw := map[string]interface{}{
		"topics":        list,
		"subscriptions": subs,
	}

	return &Yield{
		Request:     inv.Request,
		ArgumentsKw: kw,
	}, nil

}
