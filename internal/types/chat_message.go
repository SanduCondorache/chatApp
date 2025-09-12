package types

import (
	"encoding/json"
	"time"
)

type ChatMessage struct {
	Send       string    `json:"send_id"`
	Recv       string    `json:"recv_id"`
	Msg        string    `json:"msg"`
	Created_at time.Time `json:"created_at"`
}

func NewChatMessage(send, recv string, msg string, time time.Time) *ChatMessage {
	return &ChatMessage{
		Msg:        msg,
		Created_at: time,
		Send:       send,
		Recv:       recv,
	}
}

func (m *ChatMessage) ToEnvelopePayload() ([]byte, error) {
	return json.Marshal(m)
}
