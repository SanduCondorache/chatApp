package types

import "encoding/json"

type ChatMessage struct {
	From    string `json:"from"`
	Payload []byte `json:"payload"`
}

func NewChatMessage(From string, Payload []byte) *ChatMessage {
	return &ChatMessage{
		From:    From,
		Payload: Payload,
	}
}

func (m *ChatMessage) ToEnvelopePayload() ([]byte, error) {
	return json.Marshal(m)
}
