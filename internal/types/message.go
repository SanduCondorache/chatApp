package types

import "encoding/json"

type Message struct {
	Payload []byte `json:"payload"`
}

func NewMessage(payload string) *Message {
	return &Message{
		Payload: []byte(payload),
	}
}

func (m *Message) ToEnvelopePayload() ([]byte, error) {
	return json.Marshal(m)
}
