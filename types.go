package wampire

import (
	"github.com/nu7hatch/gouuid"
	"log"
)

const (
	HELLO        MsgType = 1
	WELCOME      MsgType = 2
	ABORT        MsgType = 3
	ERROR        MsgType = 10
	PUBLISH      MsgType = 11
	PUBLISHED    MsgType = 12
	SUBSCRIBE    MsgType = 13
	SUBSCRIBED   MsgType = 14
	UNSUBSCRIBE  MsgType = 15
	UNSUBSCRIBED MsgType = 16
	CALL         MsgType = 20
	RESULT       MsgType = 21
	REGISTER     MsgType = 22
	REGISTERED   MsgType = 23
	UNREGISTER   MsgType = 24
	UNREGISTERED MsgType = 25
)

type ID string
type PeerID string

func NewId() ID {
	idV4, err := uuid.NewV4()
	if err != nil {
		log.Println("error generating v4 ID:", err)
	}

	id, err := uuid.NewV5(idV4, []byte("message"))
	if err != nil {
		log.Println("error generating v5 ID:", err)
	}

	return ID(id.String())
}

type URI string

type Topic string

type MsgType int

type Message interface {
	MsgType() MsgType
}

func (t MsgType) NewMessage() Message {
	switch t {
	case HELLO:
		return new(Hello)
	case WELCOME:
		return new(Welcome)
	case ABORT:
		return new(Abort)
	case ERROR:
		return new(Error)
	case PUBLISH:
		return new(Publish)
	case PUBLISHED:
		return new(Published)
	case SUBSCRIBE:
		return new(Subscribe)
	case SUBSCRIBED:
		return new(Subscribed)
	case UNSUBSCRIBE:
		return new(Unsubscribe)
	case UNSUBSCRIBED:
		return new(Unsubscribed)
	case CALL:
		return new(Call)
	case RESULT:
		return new(Result)
	case REGISTER:
		return new(Register)
	case REGISTERED:
		return new(Registered)
	case UNREGISTER:
		return new(Unregister)
	case UNREGISTERED:
		return new(Unregistered)
	default:
		return nil
	}
}

func (t MsgType) String() string {
	switch t {
	case HELLO:
		return "HELLO"
	case WELCOME:
		return "WELCOME"
	case ABORT:
		return "ABORT"
	case ERROR:
		return "ERROR"
	case PUBLISH:
		return "PUBLISH"
	case PUBLISHED:
		return "PUBLISHED"
	case SUBSCRIBE:
		return "SUBSCRIBE"
	case SUBSCRIBED:
		return "SUBSCRIBED"
	case UNSUBSCRIBE:
		return "UNSUBSCRIBE"
	case UNSUBSCRIBED:
		return "UNSUBSCRIBED"
	case CALL:
		return "CALL"
	case RESULT:
		return "RESULT"
	case REGISTER:
		return "REGISTER"
	case REGISTERED:
		return "REGISTERED"
	case UNREGISTER:
		return "UNREGISTER"
	case UNREGISTERED:
		return "UNREGISTERED"
	default:
		panic("Invalid message type")
	}
}

// [HELLO, Details|dict]
type Hello struct {
	Id      ID
	Details map[string]interface{}
}

func (msg *Hello) MsgType() MsgType {
	return HELLO
}

// [WELCOME, Session|id, Details|dict]
type Welcome struct {
	Id      ID
	Details map[string]interface{}
}

func (msg *Welcome) MsgType() MsgType {
	return WELCOME
}

// [ABORT, Details|dict, Reason|uri]
type Abort struct {
	Id      ID
	Details map[string]interface{}
	Reason  URI
}

func (msg *Abort) MsgType() MsgType {
	return ABORT
}

// [ERROR, REQUEST.Type|int, REQUEST.Request|id, Details|dict, Error|uri]
// [ERROR, REQUEST.Type|int, REQUEST.Request|id, Details|dict, Error|uri, Arguments|list]
// [ERROR, REQUEST.Type|int, REQUEST.Request|id, Details|dict, Error|uri, Arguments|list, ArgumentsKw|dict]
type Error struct {
	Request     ID
	Error       URI
	Details     map[string]interface{}
	Arguments   []interface{}          `wamp:"omitempty"`
	ArgumentsKw map[string]interface{} `wamp:"omitempty"`
}

func (msg *Error) MsgType() MsgType {
	return ERROR
}

// [PUBLISH, Request|id, Options|dict, Topic|uri]
// [PUBLISH, Request|id, Options|dict, Topic|uri, Arguments|list]
// [PUBLISH, Request|id, Options|dict, Topic|uri, Arguments|list, ArgumentsKw|dict]
type Publish struct {
	Request     ID
	Options     map[string]interface{}
	Topic       Topic
	Arguments   []interface{}          `wamp:"omitempty"`
	ArgumentsKw map[string]interface{} `wamp:"omitempty"`
}

func (msg *Publish) MsgType() MsgType {
	return PUBLISH
}

// [PUBLISHED, PUBLISH.Request|id, Publication|id]
type Published struct {
	Request     ID
}

func (msg *Published) MsgType() MsgType {
	return PUBLISHED
}

// [SUBSCRIBE, Request|id, Options|dict, Topic|uri]
type Subscribe struct {
	Request ID
	Options map[string]interface{}
	Topic   Topic
}

func (msg *Subscribe) MsgType() MsgType {
	return SUBSCRIBE
}

// [SUBSCRIBED, SUBSCRIBE.Request|id, Subscription|id]
type Subscribed struct {
	Request      ID
	Subscription ID
}

func (msg *Subscribed) MsgType() MsgType {
	return SUBSCRIBED
}

// [UNSUBSCRIBE, Request|id, SUBSCRIBED.Subscription|id]
type Unsubscribe struct {
	Request      ID
	Subscription ID
}

func (msg *Unsubscribe) MsgType() MsgType {
	return UNSUBSCRIBE
}

// [UNSUBSCRIBED, UNSUBSCRIBE.Request|id]
type Unsubscribed struct {
	Request ID
}

func (msg *Unsubscribed) MsgType() MsgType {
	return UNSUBSCRIBED
}

// CallResult represents the result of a CALL.
type CallResult struct {
	Args   []interface{}
	Kwargs map[string]interface{}
	Error  URI
}

// [CALL, Request|id, Options|dict, Procedure|uri]
// [CALL, Request|id, Options|dict, Procedure|uri, Arguments|list]
// [CALL, Request|id, Options|dict, Procedure|uri, Arguments|list, ArgumentsKw|dict]
type Call struct {
	Request     ID
	Options     map[string]interface{}
	Procedure   URI
	Arguments   []interface{}          `wamp:"omitempty"`
	ArgumentsKw map[string]interface{} `wamp:"omitempty"`
}

func (msg *Call) MsgType() MsgType {
	return CALL
}

// [RESULT, CALL.Request|id, Details|dict]
type Result struct {
	Request     ID
	Details     map[string]interface{}
	Arguments   []interface{}          `wamp:"omitempty"`
	ArgumentsKw map[string]interface{} `wamp:"omitempty"`
}

func (msg *Result) MsgType() MsgType {
	return RESULT
}

// [REGISTER, Request|id, Options|dict, Procedure|uri]
type Register struct {
	Request   ID
	Options   map[string]interface{}
	Procedure URI
}

func (msg *Register) MsgType() MsgType {
	return REGISTER
}

// [REGISTERED, REGISTER.Request|id, Registration|id]
type Registered struct {
	Request      ID
	Registration ID
}

func (msg *Registered) MsgType() MsgType {
	return REGISTERED
}

// [UNREGISTER, Request|id, REGISTERED.Registration|id]
type Unregister struct {
	Request      ID
	Registration ID
}

func (msg *Unregister) MsgType() MsgType {
	return UNREGISTER
}

// [UNREGISTERED, UNREGISTER.Request|id]
type Unregistered struct {
	Request ID
}

func (msg *Unregistered) MsgType() MsgType {
	return UNREGISTERED
}
