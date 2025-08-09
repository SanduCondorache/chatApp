package types

import "encoding/json"

type ExitMessage struct {
	Payload []byte `json:"payload"`
}

func NewExitMessage() *ExitMessage {
	return &ExitMessage{
		Payload: []byte("exit"),
	}
}

func (m *ExitMessage) ToEnvelopePayload() ([]byte, error) {
	return json.Marshal(m)
}
