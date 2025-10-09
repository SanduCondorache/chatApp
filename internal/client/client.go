package client

import (
	"encoding/json"
	"log/slog"
	"net/url"

	"github.com/SanduCondorache/chatApp/internal/config"
	"github.com/SanduCondorache/chatApp/internal/types"
	"github.com/SanduCondorache/chatApp/utils"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	MsgCh  chan string
	ChatCh chan types.ChatMessage
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
		slog.Error("connecting to server")
		return nil, err
	}

	client := &Client{
		conn:   conn,
		MsgCh:  make(chan string, 100),
		ChatCh: make(chan types.ChatMessage, 100),
	}

	go client.readloop()
	return client, nil
}

func (c *Client) readloop() {
	for {
		msg := types.Envelope{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			slog.Error("read json", "err", err)
			close(c.MsgCh)
			return
		}

		switch msg.Type {
		case types.MsgRecv:
			var m types.ChatMessage
			if err := json.Unmarshal(msg.Payload, &m); err != nil {
				slog.Error("unmarshal error", "err", err)
				continue
			}

			c.ChatCh <- m

		case types.MsgSent:
			var m types.Message

			if err := json.Unmarshal(msg.Payload, &m); err != nil {
				slog.Error("unmarshal error", "err", err)
				continue
			}

			c.MsgCh <- "message_sent"

		case types.GetConn, types.GetMsg, types.GetChats:

			c.MsgCh <- string(msg.Payload)

		default:
			var m types.Message

			if err := json.Unmarshal(msg.Payload, &m); err != nil {
				slog.Error("unmarshal error", "err", err)
				continue
			}

			c.MsgCh <- string(m.Payload)

		}
	}
}

func (c *Client) ReadMessage() (string, error) {
	msg, ok := <-c.MsgCh

	if !ok {
		return "", types.ErrorConnectionClosed
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
