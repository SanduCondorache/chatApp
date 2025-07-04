package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/SanduCondorache/chatApp/cmd/types"
	"github.com/SanduCondorache/chatApp/utils"
)

type Server struct {
	ListenAddr string
	Ln         net.Listener
	Quitch     chan struct{}
	Msgch      chan types.Message
	Clients    map[net.Conn]bool
	AddCh      chan net.Conn
	RemoveCh   chan net.Conn
}

func NewServer(listenAddr string) *Server {
	return &Server{
		ListenAddr: listenAddr,
		Quitch:     make(chan struct{}),
		Msgch:      make(chan types.Message, 10),
		Clients:    make(map[net.Conn]bool),
		AddCh:      make(chan net.Conn),
		RemoveCh:   make(chan net.Conn),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}

	defer ln.Close()
	s.Ln = ln

	go s.acceptLoop()
	go s.broadcastLoop()
	go s.listenForExitCommand()

	<-s.Quitch
	close(s.Msgch)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.Ln.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
			continue
		}

		fmt.Printf("new connection: %s\n", utils.NormalizeAddr(conn.RemoteAddr().String()))

		go func(c net.Conn) {
			s.AddCh <- c
			s.readLoop(c)
		}(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	for {
		var msg types.Message
		if err := decoder.Decode(&msg); err != nil {
			fmt.Println("decode error: ", err)
			s.RemoveCh <- conn
			return
		}

		switch msg.Type {
		case "init":
			fmt.Println("Initial messsage from " + utils.NormalizeAddr(conn.RemoteAddr().String()))
		case "chat":
			s.Msgch <- types.Message{
				From:    utils.NormalizeAddr(conn.RemoteAddr().String()),
				Payload: bytes.TrimSpace(msg.Payload),
				Type:    msg.Type,
			}
		default:
			fmt.Println("unknown message type ", msg.Type)
		}

	}
}

func (s *Server) broadcastLoop() {
	for {
		select {
		case conn := <-s.AddCh:
			fmt.Println("Adding client:", utils.NormalizeAddr(conn.RemoteAddr().String()))
			s.Clients[conn] = true
		case conn := <-s.RemoveCh:
			delete(s.Clients, conn)
			conn.Close()
		case msg := <-s.Msgch:
			fmt.Printf("Message received from %s message: %s", utils.NormalizeAddr(msg.From), msg.Payload)
			for conn := range s.Clients {
				if utils.NormalizeAddr(conn.RemoteAddr().String()) == msg.From {
					continue
				}
				_, err := conn.Write(fmt.Appendf(nil, "%s\n", msg.Payload))

				if err != nil {
					fmt.Println("write error: ", err)
					s.RemoveCh <- conn
				}
			}
		case <-s.Quitch:
			return
		}
	}
}

func (s *Server) listenForExitCommand() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if scanner.Text() == "exit" {
			fmt.Println("Shutting down server...")
			close(s.Quitch)
			s.Ln.Close()

			for conn := range s.Clients {
				conn.Close()
			}
			return
		}
	}
}
