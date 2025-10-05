package types

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"passowrd"`
}

func NewUser(Username, Email, Password string) *User {
	return &User{
		Username: Username,
		Email:    Email,
		Password: Password,
	}
}

func (u *User) ToEnvelopePayload() ([]byte, error) {
	return json.Marshal(u)
}

func ReadUser(msg Envelope, conn *websocket.Conn) (*User, error) {
	u := &User{}
	if err := json.Unmarshal(msg.Payload, u); err != nil {
		return nil, err
	}
	return u, nil
}
