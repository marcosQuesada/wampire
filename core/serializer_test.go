package core

import (
	"testing"
)

func TestJsonSerializer(t *testing.T) {
	msg := &Hello{
		Realm:   URI("fooRealm"),
		Details: map[string]interface{}{"foo": "bar"},
	}

	s := NewJSONSerializer()
	data, err := s.Serialize(msg)
	if err != nil {
		t.Error("Unexpected error serializing ", err)
	}

	var rcvMessage Message
	rcvMessage, err = s.Deserialize(data)
	if err != nil {
		t.Error("Unexpected error deserializing ", err)
	}

	switch rcvMessage.(type) {
	case *Hello:
		h := rcvMessage.(*Hello)
		/*		if msg.Id != h.Id {
				t.Error("Message Ids don't match", msg.Id, h.Id)
			}*/

		if "bar" != h.Details["foo"] {
			t.Error("Message Payload don't match")
		}
	default:
		t.Error("Wrong type")
	}
}

// [HELLO, Details|dict]
// [WELCOME, Session|id, Details|dict]
// [ABORT, Details|dict, Reason|uri]
// [ERROR, REQUEST.Type|int, REQUEST.Request|id, Details|dict, Error|uri]
// [PUBLISH, Request|id, Options|dict, Topic|uri]
func TestMessageToList(t *testing.T) {
	hello := &Hello{
		Realm:   URI("fooUri"),
		Details: map[string]interface{}{"foo": "bar"},
	}
	e := &defaultEncoder{}
	l := e.ToList(hello)

	if len(l) != 3 {
		t.Errorf("Unexpected encoded message to list")
	}

	if l[0] != 1 {
		t.Errorf("Unexpected encoded message to list")
	}

	if l[1] != URI("fooUri") {
		t.Errorf("Unexpected encoded message to list")
	}

	d, ok := l[2].(map[string]interface{})
	if !ok {
		t.Errorf("Unexpected details type")
	}
	if d["foo"].(string) != "bar" {
		t.Errorf("Unexpected encoded message to list")
	}
}

func TestListToMessage(t *testing.T) {
	list := []interface{}{
		float64(1), URI("foo"), map[string]interface{}{"foo": "bar"},
	}
	e := &defaultEncoder{}
	hl, err := e.ToMessage(list)
	if err != nil {
		t.Error("Error converting to message ", err)
	}

	h, ok := hl.(*Hello)
	if !ok {
		t.Error("Unexpected message type")
	}

	if h.Realm != URI("foo") {
		t.Error("Unexpected message Realm")
	}

	if h.Details["foo"].(string) != "bar" {
		t.Error("Unexpected message Details")
	}
}

func TestListToMessageOnVoidFields(t *testing.T) {
	//[48,2332246451159040,{},"com.example.add2",[2,3]]
	list := []interface{}{
		float64(48), 2332246451159040, map[string]interface{}{}, URI("com.example.add2"), []interface{}{2, 3},
	}
	e := &defaultEncoder{}
	hl, err := e.ToMessage(list)
	if err != nil {
		t.Error("Error converting to message ", err)
	}

	c, ok := hl.(*Call)
	if !ok {
		t.Error("Unexpected message type")
	}

	if c.Request != 2332246451159040 {
		t.Error("Unexpected message Request ", c.Request)
	}

	if c.Procedure != URI("com.example.add2") {
		t.Error("Unexpected message Details")
	}

	if len(c.Arguments) != 2 {
		t.Error("Unexpected message Arguments")
	}

	if c.Arguments[0] != 2 {
		t.Error("Unexpected message Arguments")
	}
	if c.Arguments[1] != 3 {
		t.Error("Unexpected message Arguments")
	}
}
