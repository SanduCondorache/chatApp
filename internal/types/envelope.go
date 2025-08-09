package types

import "encoding/json"

type Envelope struct {
	Type    string `json:"type"`
	Payload json.RawMessage
}

func NewEnvelope(t string, payload json.RawMessage) *Envelope {
	return &Envelope{
		Type:    t,
		Payload: payload,
	}
}
