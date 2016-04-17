package core

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Dealer interface {
	Register(Message, *Session)
	Unregister(Message, *Session)
	Call(Message, *Session)
	Yield(Message, *Session)
	Interrupt(Message, *Session)
	Cancel(Message, *Session)
	RegisterSessionHandlers(map[URI]Handler, *inSession)
	Handlers() map[URI]Handler
}

type defaultDealer struct {
	sessionHandlers map[URI]ID
	registrations   map[ID]*Session // handler registrations from session callees
	reqListeners    *RequestListener
	mutex           *sync.RWMutex
	metaEvents      SessionMetaEventHandler
	activeTasks     map[ID]*task
}

func NewDealer(m SessionMetaEventHandler) *defaultDealer {
	d := &defaultDealer{
		sessionHandlers: make(map[URI]ID),
		registrations:   make(map[ID]*Session),
		mutex:           &sync.RWMutex{},
		reqListeners:    NewRequestListener(),
		metaEvents:      m,
		activeTasks:     make(map[ID]*task),
	}

	return d
}

func (d *defaultDealer) Register(msg Message, s *Session) {
	log.Println("Register procedure: ", msg.(*Register).Procedure)
	d.mutex.Lock()
	defer d.mutex.Unlock()

	register := msg.(*Register)
	if _, ok := d.sessionHandlers[register.Procedure]; ok {
		uri := fmt.Sprintf("%s handler already registered ", register.Procedure)
		response := &Error{
			Request: register.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
	}

	id := NewId()
	d.registrations[id] = s
	d.sessionHandlers[register.Procedure] = id
	s.addRegistration(id, register.Procedure)
	d.metaEvents.Fire(
		s.ID(),
		URI("wampire.registration.on_register"),
		map[string]interface{}{},
	)

	if s.ID() == PeerID("internal") {
		return
	}
	response := &Registered{
		Request:      register.Request,
		Registration: id,
	}
	s.Send(response)
}

func (d *defaultDealer) Unregister(msg Message, s *Session) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	unregister := msg.(*Unregister)
	_, ok := d.registrations[unregister.Registration]
	if !ok {
		uri := fmt.Sprintf("%d handler not registered ", unregister.Request)
		log.Println(uri)
		response := &Error{
			Request: unregister.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
	}

	uri, err := s.uriFromRegistration(unregister.Registration)
	if err != nil {
		uri := fmt.Sprintf("%d handler not found on Session ", unregister.Registration)
		log.Println(uri)
		response := &Error{
			Request: unregister.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
	}

	_, ok = d.sessionHandlers[uri]
	if !ok {
		uri := fmt.Sprintf("%d peerHandlers not found  %s", unregister.Registration, uri)
		log.Println(uri)
		response := &Error{
			Request: unregister.Request,
			Error:   URI(uri),
		}
		s.Send(response)
		return
	}

	//delete uri s
	delete(d.sessionHandlers, uri)
	//delete dealer registrations
	delete(d.registrations, unregister.Registration)
	//unregister uri from session
	s.unregister(uri)

	d.metaEvents.Fire(
		s.ID(),
		URI("wampire.registration.on_unregister"),
		map[string]interface{}{},
	)

	response := &Unregistered{
		Request: unregister.Request,
	}
	s.Send(response)
}

func (d *defaultDealer) Call(msg Message, s *Session) {
	call := msg.(*Call)
	registration, ok := d.sessionHandlers[call.Procedure]
	if !ok {
		uri := "Registration not found on sessionHandlers"
		log.Print(uri, msg.MsgType())
		response := &Error{
			Error: URI(uri),
		}
		s.Send(response)
		return
	}

	// Forward as invocation to peer calleee
	invocation := &Invocation{
		Request:      call.Request,
		Registration: registration,
		Details:      call.Options,
		Arguments:    call.Arguments,
		ArgumentsKw:  call.ArgumentsKw,
	}

	log.Println("Invocation:", call.Procedure, "Request is ", call.Request, "origin peer ", s.ID())
	calleeSession, ok := d.registrations[registration]
	if !ok {
		uri := "Registration not found "
		log.Print(uri, msg.MsgType())
		response := &Error{
			Error: URI(uri),
		}
		s.Send(response)
		return
	}

	// register call request in active task map
	task := newTask(s, invocation.Request, call.Procedure, false)
	d.addTask(task)

	// Handle Invocation
	err := calleeSession.do(invocation)
	if err != nil {
		log.Println("Error calleeSession do", err, invocation)
		response := &Error{
			Error: URI("calleeSession invocation do Error"),
			Details:map[string]interface{}{"error": err},
		}
		s.Send(response)
		return
	}
}

func (d *defaultDealer) Yield(msg Message, s *Session) {
	yield := msg.(*Yield)

	d.mutex.RLock()
	task, ok := d.activeTasks[yield.Request]
	d.mutex.RUnlock()
	if !ok {
		log.Println("Error, task not found! ", yield.Request)
		return
	}

	response := &Result{
		Request:     yield.Request,
		Details:     yield.Options,
		Arguments:   yield.Arguments,
		ArgumentsKw: yield.ArgumentsKw,
	}
	task.session.Send(response)

	if !task.progressive {
		d.removeTask(task)
	}
}

func (d *defaultDealer) Cancel(msg Message, s *Session) {
	cancel := msg.(*Cancel)

	log.Println("Canceling task ", cancel.Request)
	task, ok := d.activeTasks[cancel.Request]
	if !ok {
		log.Println("Current task not found! ", cancel.Request)
		return
	}

	// close task channel, don't handle response here
	close(task.terminate)

	d.removeTask(task)
}

// As cancel workaround
func (d *defaultDealer) Interrupt(msg Message, s *Session) {
	interrupt := msg.(*Interrupt)

	d.mutex.RLock()
	task, ok := d.activeTasks[interrupt.Request]
	d.mutex.RUnlock()
	if !ok {
		log.Println("Error, task not found! ", interrupt.Request)
		return
	}
	task.session.Send(interrupt)

	d.removeTask(task)

	return
}

func (d *defaultDealer) RegisterSessionHandlers(handlers map[URI]Handler, s *inSession) {
	for uri, h := range handlers {
		d.Register(&Register{Request: NewId(), Procedure: uri}, s.session)
		err := s.session.register(uri, h)
		if err != nil {
			log.Println("InSession Error registerig ", uri)
		}
	}
}

func (d *defaultDealer) Handlers() map[URI]Handler {
	return map[URI]Handler{
		"wampire.core.dealer.dump":         d.dumpDealer,
		"wampire.core.long.duration.call":  d.longDurationTask,
		"wampire.core.dealer.active.tasks": d.dumpActiveTasks,
	}
}

func (d *defaultDealer) addTask(task *task) {
	d.mutex.Lock()
	log.Println("Add invocation request ", task.request)
	d.activeTasks[task.request] = task
	d.mutex.Unlock()
}

func (d *defaultDealer) removeTask(task *task) {
	d.mutex.Lock()
	log.Println("Remove invocation request ", task.request)
	delete(d.activeTasks, task.request)
	d.mutex.Unlock()
}

func (d *defaultDealer) dumpDealer(msg Message) (Message, error) {
	d.mutex.RLock()
	sessions := d.sessionHandlers
	registrations := d.registrations
	d.mutex.RUnlock()

	list := map[string]interface{}{}
	for uri, id := range sessions {
		list[string(uri)] = id
	}

	regs := map[string]interface{}{}
	for id, s := range registrations {
		regs[fmt.Sprintf("%d", id)] = s.ID()
	}
	inv := msg.(*Invocation)
	kw := map[string]interface{}{
		"session_handlers": list,
		"registrations":    regs,
	}

	return &Yield{
		Request:     inv.Request,
		ArgumentsKw: kw,
	}, nil
}

func (d *defaultDealer) longDurationTask(msg Message) (Message, error) {
	invocation := msg.(*Invocation)
	log.Println("Invoking long duration task")
	task, ok := d.activeTasks[invocation.Request]
	if !ok {
		log.Println("Current task not found! ")
	}

	updateTickerDuration := time.Second * 50
	if v, ok := invocation.Details["receive_progress"];ok && v.(bool) {
		updateTickerDuration = time.Second * 5
	}
	updateTicker := time.NewTicker(updateTickerDuration)
	timeout := time.NewTimer(time.Second * 50)
	loopIterations := 0
	for {
		select {
		case <-task.terminate:
			log.Println("Canceled task ", invocation.Request)
			return &Interrupt{
				Request: invocation.Request,
			}, nil
		case <-updateTicker.C:
			loopIterations++
			d.mutex.RLock()
			task, ok := d.activeTasks[invocation.Request]
			d.mutex.RUnlock()
			if !ok {
				log.Println("Error, task not found! ", invocation.Request)
				continue
			}
			log.Println("Updating task ", invocation.Request)
			update := &Yield{
				Request: invocation.Request,
				ArgumentsKw: map[string]interface{}{"update": loopIterations},
			}

			task.session.Send(update)
		case <-timeout.C:
			log.Println("done long duration task done")
			d.mutex.Lock()
			task, ok := d.activeTasks[invocation.Request]
			d.mutex.Unlock()
			if !ok {
				log.Println("Error, task not found! ", invocation.Request)
				continue
			}
			task.progressive = false

			return &Yield{
				Request: invocation.Request,
				Arguments:[]interface{}{"done 100%"},
			}, nil
		}
	}
}

func (d *defaultDealer) dumpActiveTasks(msg Message) (Message, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	inv := msg.(*Invocation)
	activeTasks := []interface{}{}
	for taskID, _ := range d.activeTasks {
		if taskID != inv.Request {
			activeTasks = append(activeTasks, taskID)
		}
	}

	kw := map[string]interface{}{
		"tasks": activeTasks,
	}
	return &Yield{
		Request:     inv.Request,
		ArgumentsKw: kw,
	}, nil
}

type task struct {
	session     *Session
	request     ID
	procedure   URI
	progressive bool
	terminate   chan struct{}
}

func newTask(s *Session, id ID, uri URI, p bool) *task {
	return &task{
		session:     s,
		request:     id,
		procedure:   uri,
		progressive: p,
		terminate:   make(chan struct{}),
	}
}
