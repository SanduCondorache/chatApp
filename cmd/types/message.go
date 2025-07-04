package types

type Message struct {
	Type    string `json:"type"`
	From    string `json:"from"`
	Payload []byte `json:"payload"`
}

func NewMessage(From, Type string, Payload []byte) Message {
	return Message{
		From:    From,
		Type:    Type,
		Payload: Payload,
	}
}
