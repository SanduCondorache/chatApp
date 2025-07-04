package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/SanduCondorache/chatApp/cmd/types"
)

func readMessage(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("Server:", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Disconnected from server:", err)
	}
}

func sendMessage(conn net.Conn, payload string, t string) error {

	data, err := json.Marshal(types.NewMessage("", t, []byte(payload)))
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(data))

	return nil
}

func RunClient() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server")
		return
	}

	defer conn.Close()

	var username, email, password string

	fmt.Printf("Enter username: ")
	_, err = fmt.Scanln(&username)

	if err != nil {
		fmt.Println("Input error: ", err)
		return
	}

	fmt.Printf("Enter email: ")
	_, err = fmt.Scanln(&email)

	if err != nil {
		fmt.Println("Input error: ", err)
		return
	}

	fmt.Printf("Enter password: ")
	_, err = fmt.Scanln(&password)

	if err != nil {
		fmt.Println("Input error: ", err)
		return
	}

	payload := fmt.Sprintf("username:%s,email:%s,password:%s", username, email, password)

	err = sendMessage(conn, payload, "init")

	if err != nil {
		fmt.Println("Sending error: ", err)
		return
	}

	go readMessage(conn)

	for {
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Println("Input error: ", err)
			continue
		}

		payload := fmt.Sprintf("input:%s", input)

		err = sendMessage(conn, payload, "chat")

		if err != nil {
			fmt.Println("Sending error: ", err)
			return
		}
	}

}
