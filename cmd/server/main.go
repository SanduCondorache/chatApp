package main

import (
	"log"

	s "github.com/SanduCondorache/chatApp/cmd/server/server"
)

func main() {
	var server *s.Server
	server = s.NewServer(":8080")
	log.Fatal(server.Start())
}
