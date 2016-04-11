package core

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

type Serializer interface {
	Serialize(Message) ([]byte, error)
	Deserialize([]byte) (Message, error)
}

type Encoder interface {
	ToMessage([]interface{}) (Message, error)
	ToList(Message) []interface{}
}

type JSONSerializer struct {
	Encoder
}

func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{
		Encoder: &defaultEncoder{},
	}
}

func (s *JSONSerializer) Serialize(m Message) ([]byte, error) {
	payload := s.ToList(m)

	return json.Marshal(payload)
}

func (s *JSONSerializer) Deserialize(data []byte) (Message, error) {
	payload := []interface{}{}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}
	if len(payload) <= 1 {
		panic(payload)
	}
	return s.ToMessage(payload)
}

type defaultEncoder struct{}

func (e *defaultEncoder) ToList(msg Message) []interface{} {
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

func (e *defaultEncoder) ToMessage(l []interface{}) (Message, error) {
	msgType := MsgType(int(l[0].(float64)))
	msg := msgType.NewMessage()
	if msg == nil {
		return nil, fmt.Errorf("Unsupported message format")
	}
	val := reflect.ValueOf(msg).Elem()
	typ := reflect.TypeOf(msg).Elem()
	nl := l[1:]

	msgMap := make(map[string]interface{}, len(nl))
	for i := 0; i < val.Type().NumField(); i++ {
		//On void extra fields do nothing
		if len(nl) > i {
			msgMap[typ.Field(i).Name] = nl[i]
		}
	}

	err := mapstructure.Decode(msgMap, msg)

	return msg, err
}
