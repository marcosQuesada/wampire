package core

import (
	"encoding/json"
	"reflect"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type Serializer interface {
	Serialize(Message) ([]byte, error)
	Deserialize([]byte) (Message, error)
}

type nopSerializer struct{}

func (s *nopSerializer) Serialize(m Message) []byte {
	return nil
}

func (s *nopSerializer) Deserialize(m []byte) Message {
	return nil
}

type JsonSerializer struct{}

func (s *JsonSerializer) Serialize(m Message) ([]byte, error) {
	payload := s.toList(m)

	return json.Marshal(payload)
}

func (s *JsonSerializer) Deserialize(data []byte) (Message, error) {
	payload := []interface{}{}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}
	if len(payload) <= 1 {
		panic(payload)
	}

	return s.toMessage(payload)
}

func (s *JsonSerializer) toList(msg Message) []interface{} {
	ret := []interface{}{int(msg.MsgType())}
	val := reflect.ValueOf(msg).Elem()

	// @TODO: Handle tag annotations
	for i := 0; i < val.Type().NumField(); i++ {
		/*		tag := val.Type().Field(i).Tag.Get("wamp")
				log.Println("Tag is ", tag)
				if strings.Contains(tag, "omitempty") || val.Field(i).Len() == 0 {
					break
				}*/
		ret = append(ret, val.Field(i).Interface())
	}
	return ret
}

// @TODO: It's just handling void fields on the final part of the messages
// what's about void fields in the middle of the message??
func (s *JsonSerializer) toMessage(l []interface{}) (Message, error) {
	msgType := MsgType(int(l[0].(float64)))
	msg := msgType.NewMessage()
	if msg == nil {
		return nil, fmt.Errorf("Unsupported message format")
	}
	val := reflect.ValueOf(msg).Elem()
	typ := reflect.TypeOf(msg).Elem()
	nl := l[1:]

	msgMap := make(map[string]interface{}, len(nl))
	for i:=0; i < val.Type().NumField(); i++ {
		//On void extra fields do nothing
		if len(nl)>i {
			msgMap[typ.Field(i).Name] = nl[i]
		}
	}

	err := mapstructure.Decode(msgMap, msg)

	return msg, err
}