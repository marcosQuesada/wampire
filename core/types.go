package core

import (
	"sync/atomic"
"log"
"github.com/nu7hatch/gouuid"
)

// https://tools.ietf.org/html/draft-oberstet-hybi-tavendo-wamp-02

const (
	HELLO        MsgType = 1
	WELCOME      MsgType = 2
	ABORT        MsgType = 3
	CHALLENGE    MsgType = 4
	AUTHENTICATE MsgType = 5
	GOODBYE      MsgType = 6
	ERROR        MsgType = 8
	PUBLISH      MsgType = 16
	PUBLISHED    MsgType = 17
	SUBSCRIBE    MsgType = 32
	SUBSCRIBED   MsgType = 33
	UNSUBSCRIBE  MsgType = 34
	UNSUBSCRIBED MsgType = 35
	EVENT        MsgType = 36
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

func NewStringId() PeerID {
	idV4, err := uuid.NewV4()
	if err != nil {
		log.Println("error generating v4 ID:", err)
	}

	id, err := uuid.NewV5(idV4, []byte("message"))
	if err != nil {
		log.Println("error generating v5 ID:", err)
	}

	return PeerID(id.String())
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
	case CHALLENGE:
		return new(Challenge)
	case AUTHENTICATE:
		return new(Authenticate)
	case GOODBYE:
		return new(Goodbye)
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
	case EVENT:
		return new(Event)
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
	case INVOCATION:
		return new(Invocation)
	case INTERRUPT:
		return new(Interrupt)
	case YIELD:
		return new(Yield)
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
	case EVENT:
		return "EVENT"
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
	case INVOCATION:
		return "INVOCATION"
	case INTERRUPT:
		return "INTERRUPT"
	case YIELD:
		return "YIELD"
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

// [CHALLENGE, AuthMethod|string, Extra|dict]
type Challenge struct {
	AuthMethod string
	Extra      map[string]interface{}
}

func (msg *Challenge) MsgType() MsgType {
	return CHALLENGE
}

// [AUTHENTICATE, Signature|string, Extra|dict]
type Authenticate struct {
	Signature string
	Extra     map[string]interface{}
}

func (msg *Authenticate) MsgType() MsgType {
	return AUTHENTICATE
}

// [GOODBYE, Details|dict, Reason|uri]
type Goodbye struct {
	Details map[string]interface{}
	Reason  URI
}

func (msg *Goodbye) MsgType() MsgType {
	return GOODBYE
}

// [8, 66, 788923562, {}, "wamp.error.no_such_registration"]
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

// [EVENT, SUBSCRIBED.Subscription|id, PUBLISHED.Publication|id, Details|dict]
// [EVENT, SUBSCRIBED.Subscription|id, PUBLISHED.Publication|id, Details|dict, PUBLISH.Arguments|list]
// [EVENT, SUBSCRIBED.Subscription|id, PUBLISHED.Publication|id, Details|dict, PUBLISH.Arguments|list,
//     PUBLISH.ArgumentsKw|dict]
type Event struct {
	Subscription ID
	Publication  ID
	Details      map[string]interface{}
	Arguments    []interface{}          `wamp:"omitempty"`
	ArgumentsKw  map[string]interface{} `wamp:"omitempty"`
}

func (msg *Event) MsgType() MsgType {
	return EVENT
}

// [48, 7814135, {}, "com.myapp.echo", ["Hello, world!"]]
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

// [68, 6131533, 9823527, {}, ["Hello, world!"]]
// [INVOCATION, Request|id, REGISTERED.Registration|id, Details|dict]
// [INVOCATION, Request|id, REGISTERED.Registration|id, Details|dict, CALL.Arguments|list]
// [INVOCATION, Request|id, REGISTERED.Registration|id, Details|dict, CALL.Arguments|list, CALL.ArgumentsKw|dict]
type Invocation struct {
	Request      ID
	Registration ID
	Details      map[string]interface{}
	Arguments    []interface{}          `wamp:"omitempty"`
	ArgumentsKw  map[string]interface{} `wamp:"omitempty"`
}

func (msg *Invocation) MsgType() MsgType {
	return INVOCATION
}

// [70, 6131533, {}, ["Hello, world!"]]
// [YIELD, INVOCATION.Request|id, Options|dict]
// [YIELD, INVOCATION.Request|id, Options|dict, Arguments|list]
// [YIELD, INVOCATION.Request|id, Options|dict, Arguments|list, ArgumentsKw|dict]
type Yield struct {
	Request     ID
	Options     map[string]interface{}
	Arguments   []interface{}          `wamp:"omitempty"`
	ArgumentsKw map[string]interface{} `wamp:"omitempty"`
}

func (msg *Yield) MsgType() MsgType {
	return YIELD
}

// [CANCEL, CALL.Request|id, Options|dict]
type Cancel struct {
	Request ID
	Options map[string]interface{}
}

func (msg *Cancel) MsgType() MsgType {
	return CANCEL
}

// [INTERRUPT, INVOCATION.Request|id, Options|dict]
type Interrupt struct {
	Request ID
	Options map[string]interface{}
}

func (msg *Interrupt) MsgType() MsgType {
	return INTERRUPT
}