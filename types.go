package wampire

import (
	"sync/atomic"
)

// https://tools.ietf.org/html/draft-oberstet-hybi-tavendo-wamp-02

const (
	HELLO        MsgType = 1
	WELCOME      MsgType = 2
	ABORT        MsgType = 3
	ERROR        MsgType = 8
	PUBLISH      MsgType = 16
	PUBLISHED    MsgType = 17
	SUBSCRIBE    MsgType = 32
	SUBSCRIBED   MsgType = 33
	UNSUBSCRIBE  MsgType = 34
	UNSUBSCRIBED MsgType = 35
	CALL         MsgType = 48
	CANCEL       MsgType = 49
	RESULT       MsgType = 50
	REGISTER     MsgType = 64
	REGISTERED   MsgType = 65
	UNREGISTER   MsgType = 66
	UNREGISTERED MsgType = 67
	INVOCATION   MsgType = 68
	INTERRUPT    MsgType = 69
	YIELD        MsgType = 70
)

type ID uint64
type PeerID string

// atomic counter as message IDs
var lastId uint64 = 0
func NewId() ID {
	atomic.AddUint64(&lastId, 1)

	return ID(atomic.LoadUint64(&lastId))
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

/**
       [1, "somerealm", {
         "roles": {
             "publisher": {},
             "subscriber": {}
         }
       }]
 */
/**
 Raw Wamp Hello Message
[1, "somerealm", {
	"roles": {
		"caller": {
			"features": {
				"caller_identification": true,
				"progressive_call_results": true
			}
		},
		"callee": {
			"features": {
				"progressive_call_results": true
			}
		},
		"publisher": {
			"features": {
				"subscriber_blackwhite_listing": true,
				"publisher_exclusion": true,
				"publisher_identification": true
			}
		},
		"subscriber": {
			"features": {
				"publisher_identification": true
			}
		}
	}
}]

*/
// [HELLO, Details|dict]
type Hello struct {
	Realm   URI
	Details map[string]interface{}
}

func (msg *Hello) MsgType() MsgType {
	return HELLO
}

/**
       [2, 9129137332, {
          "roles": {
             "broker": {}
          }
       }]
 */
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
	Request ID
}

func (msg *Published) MsgType() MsgType {
	return PUBLISHED
}

/**
   [32, 713845233, {}, "com.myapp.mytopic1"]
 */
// [SUBSCRIBE, Request|id, Options|dict, Topic|uri]
type Subscribe struct {
	Request ID
	Options map[string]interface{}
	Topic   Topic
}

func (msg *Subscribe) MsgType() MsgType {
	return SUBSCRIBE
}

/**
    [33, 713845233, 5512315355]
 */
// [SUBSCRIBED, SUBSCRIBE.Request|id, Subscription|id]
type Subscribed struct {
	Request      ID
	Subscription ID
}

func (msg *Subscribed) MsgType() MsgType {
	return SUBSCRIBED
}

/**
      [34, 85346237, 5512315355]
 */
// [UNSUBSCRIBE, Request|id, SUBSCRIBED.Subscription|id]
type Unsubscribe struct {
	Request      ID
	Subscription ID
}

func (msg *Unsubscribe) MsgType() MsgType {
	return UNSUBSCRIBE
}

/**
    [35, 85346237]
 */
// [UNSUBSCRIBED, UNSUBSCRIBE.Request|id]
type Unsubscribed struct {
	Request ID
}

func (msg *Unsubscribed) MsgType() MsgType {
	return UNSUBSCRIBED
}

// [CALL, Request|id, Options|dict, Procedure|uri]
// [CALL, Request|id, Options|dict, Procedure|uri, Arguments|list]
// [CALL, Request|id, Options|dict, Procedure|uri, Arguments|list, ArgumentsKw|dict]
type Call struct {
	Request     ID
	Procedure   URI
	Options     map[string]interface{}
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
	Error       URI
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
