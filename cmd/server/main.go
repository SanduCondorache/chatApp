package main

import "github.com/SanduCondorache/chatApp/internal/server"

func main() {
	s := server.NewServer(":8080")
	if err := s.Start(); err != nil {
		panic(err)
	}
}
