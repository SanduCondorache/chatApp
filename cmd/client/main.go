package main

import (
	"fmt"
	"net"

	"github.com/SanduCondorache/chatApp/cmd/client/client"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server")
		return
	}

	defer conn.Close()

	go client.ReadMessage(conn)

	for {
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Println("Input error: ", err)
			continue
		}

		_, err = conn.Write([]byte(input + "\n"))
		if err != nil {
			fmt.Println("Write error: ", err)
			return
		}
	}

}
