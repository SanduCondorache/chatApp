package client

import (
	"fmt"
	"log"
	"net/url"

	"github.com/SanduCondorache/chatApp/internal/types"
	"github.com/gorilla/websocket"
)

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

		if t == "exit" {
			log.Println("Server requested exit. Closing client...")
			close(done)
			return
		}
	}
}

func sendMessage(conn *websocket.Conn, payload types.Payload, t string) error {
	p, err := payload.ToEnvelopePayload()
	if err != nil {
		return err
	}
	data := types.NewEnvelope(t, p)
	return conn.WriteJSON(data)
}

func RunClient() {
	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:8080",
		Path:   "/ws",
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("Error connecting to server")
		return
	}
	defer conn.Close()

	var username, email, password string

	fmt.Printf("Enter username: ")
	if _, err = fmt.Scanln(&username); err != nil {
		fmt.Println("Input error: ", err)
		return
	}

	fmt.Printf("Enter email: ")
	if _, err = fmt.Scanln(&email); err != nil {
		fmt.Println("Input error: ", err)
		return
	}

	fmt.Printf("Enter password: ")
	if _, err = fmt.Scanln(&password); err != nil {
		fmt.Println("Input error: ", err)
		return
	}

	payload := types.NewUser(username, email, password)
	if err = sendMessage(conn, payload, "init"); err != nil {
		fmt.Println("Sending error: ", err)
		return
	}

	done := make(chan struct{})
	go readMessage(conn, done)

	inputCh := make(chan string)

	go func() {
		var input string
		for {
			_, err := fmt.Scanln(&input)
			if err != nil {
				continue
			}
			inputCh <- input
		}
	}()

	for {
		select {
		case <-done:
			log.Println("Shutting down client...")
			return
		case input := <-inputCh:
			payload := types.NewMessage("", []byte(input))
			if err = sendMessage(conn, payload, "chat"); err != nil {
				fmt.Println("Sending error: ", err)
				return
			}
		}
	}
}
