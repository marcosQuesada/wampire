package core

import (
	"fmt"
	"log"
	"sync"
)

type Broker interface {
	Subscribe(Message, *Session)
	UnSubscribe(Message, *Session)
	Publish(Message, *Session)
	Handlers() map[URI]Handler
}

type defaultBroker struct {
	topics        map[Topic]map[ID]bool   //maps topics to subscriptions
	subscriptions map[ID]*Session         //a peer may have many subscriptions
	topicPeers    map[Topic]map[PeerID]ID //maps peers by topic on subscription
	mutex         *sync.RWMutex
	metaEvents    SessionMetaEventHandler
}

func NewBroker(smeh SessionMetaEventHandler) *defaultBroker {
	b := &defaultBroker{
		topics:        make(map[Topic]map[ID]bool),
		subscriptions: make(map[ID]*Session),
		topicPeers:    make(map[Topic]map[PeerID]ID),
		mutex:         &sync.RWMutex{},
		metaEvents:    smeh,
	}

	// intialize session meta event topic
	eventsTopic := Topic("wampire.session.meta.events")
	b.topics[eventsTopic] = map[ID]bool{}
	b.topicPeers[eventsTopic] = map[PeerID]ID{}
	log.Println("Session meta events topic created: wampire.session.meta.events")

	return b
}

func (b *defaultBroker) Subscribe(msg Message, s *Session) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	subscribe, ok := msg.(*Subscribe)
	log.Println("subscribe invoked ", subscribe.Topic)
	if !ok {
		log.Fatal("Unexpected type on subscribe ", msg.MsgType())
		panic("Unexpected type on subscribe")
	}

	if _, ok = b.topics[subscribe.Topic]; !ok {
		//create topic!
		b.topics[subscribe.Topic] = make(map[ID]bool)
		b.topicPeers[subscribe.Topic] = make(map[PeerID]ID)

		b.metaEvents.Fire(
			s.ID(),
			URI("wampire.subscription.on_create"),
			map[string]interface{}{},
		)
	}

	//check if subscriptor is already register to topic
	if subs, ok := b.topicPeers[subscribe.Topic][s.ID()]; ok {
		log.Println("Session already subscribed on subscription ", subs)
		response := &Error{
			Error: URI("Peer already subscribed on subscription"),
		}
		s.Send(response)
		return
	}

	subscriptionId := NewId()
	b.topicPeers[subscribe.Topic][s.ID()] = subscriptionId
	b.topics[subscribe.Topic][subscriptionId] = true
	b.subscriptions[subscriptionId] = s

	// Add subscription to session
	s.addSubscription(subscriptionId, subscribe.Topic)
	b.metaEvents.Fire(
		s.ID(),
		URI("wampire.subscription.on_subscribe"),
		map[string]interface{}{},
	)

	response := &Subscribed{
		Request:      subscribe.Request,
		Subscription: subscriptionId,
	}
	s.Send(response)
}

func (b *defaultBroker) UnSubscribe(msg Message, s *Session) {
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

		response := &Error{
			Request: unsubscribe.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
	}

	session, ok := b.subscriptions[unsubscribe.Subscription]
	if !ok {
		uri := "peer not found to this Subscription"
		log.Println(uri, unsubscribe)

		response := &Error{
			Request: unsubscribe.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
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
	if len(b.topics[topic]) == 0 && topic != Topic("wampire.session.meta.events") {
		delete(b.topics, topic)
	}
	//if void topic remove it
	if len(b.topicPeers[topic]) == 0 && topic != Topic("wampire.session.meta.events") {
		delete(b.topicPeers, topic)
		b.metaEvents.Fire(
			session.ID(),
			URI("wampire.subscription.on_delete"),
			map[string]interface{}{},
		)
	}
	b.metaEvents.Fire(
		session.ID(),
		URI("wampire.subscription.on_unsubscribe"),
		map[string]interface{}{},
	)

	response := &Unsubscribed{
		Request: unsubscribe.Request,
	}

	s.Send(response)
}

func (b *defaultBroker) Publish(msg Message, s *Session) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	publish, ok := msg.(*Publish)
	if !ok {
		log.Fatal("Unexpected type on publish ", msg.MsgType())
	}

	subscribers, ok := b.topics[publish.Topic]
	if !ok {
		uri := "Topic not found"
		log.Println(uri, publish)

		response := &Error{
			Request: publish.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
	}
	//iterate on topic subscribers
	for subscriptionId, _ := range subscribers {
		//find peer from subscriber
		session, ok := b.subscriptions[subscriptionId]
		if !ok {
			log.Println("Peer not found")
			continue
		}

		if session.ID() != s.ID() {
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
			go session.Send(event)
		}
	}

	response := &Published{
		Request: publish.Request,
	}
	s.Send(response)
}

func (b *defaultBroker) Handlers() map[URI]Handler {
	return map[URI]Handler{
		"wampire.subscription.list_subscribers":       b.listSubscribers,
		"wampire.subscription.list_topics":            b.listTopics,
		"wampire.subscription.count_subscribers":      b.countSubscribers,
		"wampire.subscription.list_topic_subscribers": b.listTopicSubscriptions,
	}
}

func (b *defaultBroker) listSubscribers(msg Message) (Message, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	subs := map[string]interface{}{}
	for id, s := range b.subscriptions {
		subs[fmt.Sprintf("%d", id)] = s.ID()
	}
	inv := msg.(*Invocation)

	return &Yield{
		Request:     inv.Request,
		ArgumentsKw: map[string]interface{}{"subscriptions": subs},
	}, nil
}

func (b *defaultBroker) listTopics(msg Message) (Message, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	topicList := []interface{}{}
	for topic, _ := range b.topics {
		topicList = append(topicList, topic)
	}

	inv := msg.(*Invocation)
	kw := map[string]interface{}{
		"topics": topicList,
	}

	return &Yield{
		Request:     inv.Request,
		ArgumentsKw: kw,
	}, nil
}

func (b *defaultBroker) countSubscribers(msg Message) (Message, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	inv := msg.(*Invocation)

	return &Yield{
		Request:   inv.Request,
		Arguments: []interface{}{len(b.subscriptions)},
	}, nil
}

func (b *defaultBroker) listTopicSubscriptions(msg Message) (Message, error) {
	inv := msg.(*Invocation)
	log.Println("Getting session ", inv.Arguments)
	if len(inv.Arguments) < 1 {
		error := "Void ID argument on list topic subscriptions"
		log.Println(error)
		return nil, fmt.Errorf(error)
	}
	topic := Topic(inv.Arguments[0].(string))
	b.mutex.RLock()
	subscribers, ok := b.topics[topic]
	b.mutex.RUnlock()
	if !ok {
		uri := fmt.Sprintf("Topic %s not found", topic)
		log.Println(uri, uri)

		return nil, fmt.Errorf("%s")
	}

	list := []interface{}{}
	for id, _ := range subscribers {
		b.mutex.RLock()
		s, ok := b.subscriptions[id]
		b.mutex.RUnlock()
		if !ok {
			uri := fmt.Sprintf("Topic %s Subscriptior %d not found", topic, id)
			log.Println(uri, uri)

			continue
		}
		list = append(list, s.ID())
	}

	return &Yield{
		Request:   inv.Request,
		Arguments: list,
	}, nil
}
