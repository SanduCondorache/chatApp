package types

import "encoding/json"

type Message struct {
	From    string `json:"from"`
	Payload []byte `json:"payload"`
}

func NewMessage(From string, Payload []byte) *Message {
	return &Message{
		From:    From,
		Payload: Payload,
	}
}

func (m *Message) ToEnvelopePayload() ([]byte, error) {
	return json.Marshal(m)
}
