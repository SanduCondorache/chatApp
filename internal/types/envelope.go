package types

import (
	"encoding/json"
)

type Envelope struct {
	Type    MessageType `json:"type"`
	Payload json.RawMessage
}

func NewEnvelope(t MessageType, payload json.RawMessage) *Envelope {
	return &Envelope{
		Type:    t,
		Payload: payload,
	}
}
