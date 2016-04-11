package core

import(
	"log"
	"fmt"
)

func (r *DefaultRouter) Handlers() map[URI]Handler {
	return map[URI]Handler{
		"wampire.session.list":  r.listSessions,
		"wampire.session.count": r.countSessions,
		"wampire.session.get":   r.getSession,
	}
}

func (r *DefaultRouter) listSessions(msg Message) (Message, error) {
	r.mutex.RLock()
	sessions := r.sessions
	r.mutex.RUnlock()

	list := []interface{}{}
	for peerId, _ := range sessions {
		list = append(list, peerId)
	}

	inv := msg.(*Invocation)

	return &Yield{
		Request:   inv.Request,
		Arguments: list,
	}, nil

}

func (r *DefaultRouter) countSessions(msg Message) (Message, error) {
	r.mutex.RLock()
	total := len(r.sessions)
	r.mutex.RUnlock()
	inv := msg.(*Invocation)

	return &Yield{
		Request:   inv.Request,
		Arguments: []interface{}{total},
	}, nil
}

func (r *DefaultRouter) getSession(msg Message) (Message, error) {
	inv := msg.(*Invocation)
	if len(inv.Arguments) < 1 {
		error := "Void ID argument on get session"
		log.Println(error)
		return nil, fmt.Errorf(error)
	}

	r.mutex.RLock()
	s, ok := r.sessions[PeerID(inv.Arguments[0].(string))]
	r.mutex.RUnlock()

	if !ok {
		error := "Router session ID %d not found "
		log.Println(error)
		return nil, fmt.Errorf(error)
	}

	subs := map[string]interface{}{}
	for id, topic := range s.getSubscriptions() {
		subs[fmt.Sprintf("%d", id)] = topic
	}
	regs := map[string]interface{}{}
	for id, uri := range s.getRegistrations() {
		regs[fmt.Sprintf("%d", id)] = uri
	}
	kw := map[string]interface{}{
		"subscriptions": subs,
		"registrations": regs,
		"initTs":        s.initTs,
	}

	return &Yield{
		Request:     inv.Request,
		ArgumentsKw: kw,
	}, nil
}
