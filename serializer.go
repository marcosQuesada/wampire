package wamp

import (
	"encoding/json"
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
	t := m.MsgType()

	data := map[string]interface{}{"type": t, "msg": m}

	return json.Marshal(&data)
}

func (s *JsonSerializer) Deserialize(data []byte) (Message, error) {
	payload := map[string]interface{}{}
	err := json.Unmarshal(data, &payload)

	msg := MsgType(int(payload["type"].(float64))).NewMessage()
	err = mapstructure.Decode(payload["msg"], msg)

	return msg, err
}
