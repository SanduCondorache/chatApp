package client

import (
	"bufio"
	"fmt"
	"net"
)

func ReadMessage(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("Server:", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Disconnected from server:", err)
	}
}
