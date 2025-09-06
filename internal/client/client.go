package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/SanduCondorache/chatApp/internal/config"
	"github.com/SanduCondorache/chatApp/internal/types"
	"github.com/gorilla/websocket"
)

func readUser() (*types.User, error) {
	var username, email, password string

	fmt.Printf("Enter username: ")
	if _, err := fmt.Scanln(&username); err != nil {
		fmt.Println("Input error: ", err)
		return nil, err
	}

	fmt.Printf("Enter email: ")
	if _, err := fmt.Scanln(&email); err != nil {
		fmt.Println("Input error: ", err)
		return nil, err
	}

	fmt.Printf("Enter password: ")
	if _, err := fmt.Scanln(&password); err != nil {
		fmt.Println("Input error: ", err)
		return nil, err
	}

	payload := types.NewUser(username, email, password)
	return payload, nil
}

func readMessage(conn *websocket.Conn, done chan struct{}) {
	var msg types.Envelope
	for {
		msg = types.Envelope{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("read error:", err)
			close(done)
			return
		}

		t := msg.Type

		switch t {
		case types.Exit:
			log.Println("Server requested exit. Closing client...")
			close(done)
			return
		}
	}
}

func ReadMessage(conn *websocket.Conn) (string, error) {
	msg := types.Envelope{}
	err := conn.ReadJSON(&msg)
	if err != nil {
		return "", err
	}
	var m types.Message

	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		return "", err
	}

	fmt.Println(string(m.Payload))

	return string(m.Payload), nil
}

func SendMessage(conn *websocket.Conn, payload types.Payload, t types.MessageType) error {
	p, err := payload.ToEnvelopePayload()
	if err != nil {
		return err
	}
	data := types.NewEnvelope(t, p)
	return conn.WriteJSON(data)
}

func auth(conn *websocket.Conn, done chan struct{}) bool {
	for {
		select {
		case <-done:
			return false
		default:
			user, err := readUser()
			if err != nil {
				log.Println("scan error: ", err)
				return false
			}
			err = SendMessage(conn, user, types.Login)
			if err != nil {
				log.Println("sending error: ", err)
				return false
			}
			msg := types.Envelope{}
			err = conn.ReadJSON(&msg)
			if err != nil {
				log.Println("read error:", err)
				close(done)
				return false
			}
			var m types.Message

			if err := json.Unmarshal(msg.Payload, &m); err != nil {
				log.Println("unmarshal error: ", err)
			}

			er := string(m.Payload)

			switch er {
			case "ok":
				return true
			case "username_taken":
				fmt.Println("Username is taken")
			}
		}
	}
}

func handleMessageSent(text string) types.Payload {
	payload := types.NewChatMessage("", []byte(text))
	return payload
}

func CreateConn() (*websocket.Conn, error) {
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
	return conn, nil
}

func RunClient() {
	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:" + config.Envs.Port,
		Path:   "/ws",
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("Error connecting to server")
		return
	}
	defer conn.Close()

	done := make(chan struct{})

	if !auth(conn, done) {
		return
	}

	go readMessage(conn, done)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-done:
			fmt.Println("Client shutting down...")
			return
		default:
			if !scanner.Scan() {
				return
			}
			text := strings.TrimSpace(scanner.Text())
			if text == "" {
				continue
			}
			payload := handleMessageSent(text)
			if err := SendMessage(conn, payload, types.Chat); err != nil {
				log.Println("Send error:", err)
				return
			}
		}
	}
}
