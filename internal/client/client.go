package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/SanduCondorache/chatApp/internal/config"
	"github.com/SanduCondorache/chatApp/internal/types"
	"github.com/SanduCondorache/chatApp/utils"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn  *websocket.Conn
	MsgCh chan string
}

func NewClient() (*Client, error) {
	utils.InitLogger()
	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:" + config.Envs.Port,
		Path:   "/ws",
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("Error connecting to server")
		return nil, err
	}

	client := &Client{
		conn:  conn,
		MsgCh: make(chan string, 100),
	}

	go client.readloop()
	return client, nil
}

func (c *Client) readloop() {
	for {
		msg := types.Envelope{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("read json error:", err)
			close(c.MsgCh)
			return
		}

		switch msg.Type {
		case types.MsgRecv:
			var m types.ChatMessage
			if err := json.Unmarshal(msg.Payload, &m); err != nil {
				log.Println("unmarshal error:", err)
				continue
			}

			c.MsgCh <- string(msg.Payload)
			fmt.Println(m, "loh")

		case types.MsgSent:
			var m types.Message

			if err := json.Unmarshal(msg.Payload, &m); err != nil {
				log.Println("unmarshal error:", err)
				continue
			}

			c.MsgCh <- "message_sent"

			fmt.Println(string(m.Payload), msg.Type)

		case types.GetConn, types.GetMsg:

			c.MsgCh <- string(msg.Payload)

		default:
			var m types.Message

			if err := json.Unmarshal(msg.Payload, &m); err != nil {
				log.Println("unmarshal error:", err)
				continue
			}

			fmt.Println(string(m.Payload), msg.Type)
			c.MsgCh <- string(m.Payload)

		}
	}
}

func (c *Client) ReadMessage() (string, error) {
	msg, ok := <-c.MsgCh

	if !ok {
		return "", errors.New("connection closed")
	}

	return msg, nil
}

func (c *Client) SendMessage(payload types.Payload, t types.MessageType) error {
	p, err := payload.ToEnvelopePayload()
	if err != nil {
		return err
	}
	data := types.NewEnvelope(t, p)
	return c.conn.WriteJSON(data)
}

//
// func ReadMessage(conn *websocket.Conn) (string, error) {
// 	msg := types.Envelope{}
// 	err := conn.ReadJSON(&msg)
// 	fmt.Println(msg.Type)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	switch msg.Type {
// 	case types.MsgRecv:
// 		var m types.ChatMessage
// 		if err := json.Unmarshal(msg.Payload, &m); err != nil {
// 			return "", err
// 		}
//
// 		fmt.Println(m, "loh")
//
// 		return "message_received", nil
// 	case types.MsgSent:
// 		var m types.Message
//
// 		if err := json.Unmarshal(msg.Payload, &m); err != nil {
// 			return "", err
// 		}
//
// 		fmt.Println(string(m.Payload), msg.Type)
//
// 		return "message_sent", nil
// 	default:
// 		var m types.Message
//
// 		if err := json.Unmarshal(msg.Payload, &m); err != nil {
// 			return "", err
// 		}
//
// 		fmt.Println(string(m.Payload), msg.Type)
//
// 		return string(m.Payload), nil
// 	}
// }
