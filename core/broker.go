package core

import (
	"log"
	"sync"
)

type Broker interface {
	Subscribe(Message, Peer) Message
	UnSubscribe(Message, Peer) Message
	Publish(Message, Peer) Message
}

type defaultBroker struct {
	topics              map[Topic]map[ID]bool   //maps topics to subscriptions
	subscriptions       map[ID]Peer             //a peer may have many subscriptions
	topicPeers          map[Topic]map[PeerID]ID //maps peers by topic on subscription
	topicBySubscription map[ID]Topic            // topic from Subscription ID
	mutex               *sync.RWMutex
}

func NewBroker() *defaultBroker {
	return &defaultBroker{
		topics:              make(map[Topic]map[ID]bool),
		subscriptions:       make(map[ID]Peer),
		topicPeers:          make(map[Topic]map[PeerID]ID),
		topicBySubscription: make(map[ID]Topic),
		mutex:               &sync.RWMutex{},
	}
}

func (b *defaultBroker) Subscribe(msg Message, p Peer) Message {
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
	if subs, ok := b.topicPeers[subscribe.Topic][p.ID()]; ok {
		log.Println("Peer already subscribed on subscription ", subs)
		return &Error{
			Error: URI("Peer already subscribed on subscription"),
		}
	}

	subscriptionId := NewId()
	b.topicPeers[subscribe.Topic][p.ID()] = subscriptionId
	b.topics[subscribe.Topic][subscriptionId] = true
	b.subscriptions[subscriptionId] = p
	b.topicBySubscription[subscriptionId] = subscribe.Topic

	return &Subscribed{
		Request:      subscribe.Request,
		Subscription: subscriptionId,
	}
}

func (b *defaultBroker) UnSubscribe(msg Message, p Peer) Message {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	unsubscribe, ok := msg.(*Unsubscribe)
	if !ok {
		log.Fatal("Unexpected type on UnSubscribe ", msg.MsgType())
		panic("Unexpected type on UnSubscribe")
	}

	topic, ok := b.topicBySubscription[unsubscribe.Subscription]
	if !ok {
		uri := "topic not found to this Subscription"
		log.Println(uri, unsubscribe)

		return &Error{
			Request: unsubscribe.Request,
			Error:   URI(uri),
		}
	}

	peer, ok := b.subscriptions[unsubscribe.Subscription]
	if !ok {
		uri := "peer not found to this Subscription"
		log.Println(uri, unsubscribe)

		return &Error{
			Request: unsubscribe.Request,
			Error:   URI(uri),
		}
	}

	//remove topic by subscription
	delete(b.topicBySubscription, unsubscribe.Subscription)
	//remove peer from subscription map
	delete(b.subscriptions, unsubscribe.Subscription)
	//remove peer from topic map
	delete(b.topicPeers[topic], peer.ID())
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

func (b *defaultBroker) Publish(msg Message, p Peer) Message {
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
			//send message in a non blocking way
			go peer.Send(msg)
		}
	}

	return &Published{
		Request: publish.Request,
	}
}
