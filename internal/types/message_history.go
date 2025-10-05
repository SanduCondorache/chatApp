package types

import (
	"encoding/json"
	"time"
)

type MessageHist struct {
	Direction string    `json:"direction"`
	Content   string    `json:"content"`
	Time      time.Time `json:"time"`
}

func NewMessaageHist(direction string, content string, time time.Time) *MessageHist {
	return &MessageHist{
		Direction: direction,
		Content:   content,
		Time:      time,
	}
}

func (m *MessageHist) ToEnvelopePayload() ([]byte, error) {
	return json.Marshal(m)
}
