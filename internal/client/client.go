package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

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

func handlingErrors(conn *websocket.Conn, msg types.Envelope, pauseCh chan bool) {
	p := msg.Payload
	var m types.Message

	if err := json.Unmarshal(p, &m); err != nil {
		log.Println("unmarshal error: ", err)
	}

	er := string(m.Payload)
	fmt.Println(er)

	switch er {
	case "username_taken":
		pauseCh <- true
		payload, err := readUser()
		if err != nil {
			fmt.Println("mata")
			pauseCh <- false
			return
		}

		fmt.Println(payload)

		if err = sendMessage(conn, payload, "init"); err != nil {
			fmt.Println("Sending error: ", err)
			return
		}
		pauseCh <- false

	}

}

func readMessage(conn *websocket.Conn, done chan struct{}, pauseCh chan bool) {
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

		fmt.Println(t)

		switch t {
		case "exit":
			log.Println("Server requested exit. Closing client...")
			close(done)
			return
		case "error":
			handlingErrors(conn, msg, pauseCh)
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

	payload, err := readUser()
	if err != nil {
		return
	}

	if err = sendMessage(conn, payload, "init"); err != nil {
		fmt.Println("Sending error: ", err)
		return
	}

	done := make(chan struct{})
	pauseCh := make(chan bool)
	go readMessage(conn, done, pauseCh)

	inputCh := make(chan string)

	go func() {
		var input string
		paused := false
		for {
			if paused {
				// Wait until resume
				p := <-pauseCh
				paused = p
				continue
			}

			select {
			case p := <-pauseCh:
				paused = p
				continue
			default:
				// Only read when not paused
				_, err := fmt.Scanln(&input)
				if err != nil {
					continue
				}
				inputCh <- input
			}
		}
	}()

	for {
		select {
		case <-done:
			log.Println("Shutting down client...")
			return
		case input := <-inputCh:
			payload := types.NewChatMessage("", []byte(input))
			if err = sendMessage(conn, payload, "chat"); err != nil {
				fmt.Println("Sending error: ", err)
				return
			}
		}
	}
}
