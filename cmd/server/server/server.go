package server

import (
	"bytes"
	"fmt"
	"net"

	"github.com/SanduCondorache/chatApp/utils"
)

type Message struct {
	From    string
	Payload []byte
}

type Server struct {
	ListenAddr string
	Ln         net.Listener
	Quitch     chan struct{}
	Msgch      chan Message
	Clients    map[net.Conn]bool
	AddCh      chan net.Conn
	RemoveCh   chan net.Conn
}

func NewServer(listenAddr string) *Server {
	return &Server{
		ListenAddr: listenAddr,
		Quitch:     make(chan struct{}),
		Msgch:      make(chan Message, 10),
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
	buff := make([]byte, 2048)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println("read error: ", err)
			s.RemoveCh <- conn
			return
		}

		s.Msgch <- Message{
			From:    utils.NormalizeAddr(conn.RemoteAddr().String()),
			Payload: bytes.TrimSpace(buff[:n]),
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
				if conn.RemoteAddr().String() == msg.From {
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
