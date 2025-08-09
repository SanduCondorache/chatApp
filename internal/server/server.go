package server

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	db "github.com/SanduCondorache/chatApp/internal/database"
	"github.com/SanduCondorache/chatApp/internal/types"
	"github.com/SanduCondorache/chatApp/utils"
	"github.com/gorilla/websocket"
)

type Server struct {
	ListenAddr string
	Upgrader   websocket.Upgrader
	Clients    map[*websocket.Conn]bool
	AddCh      chan *websocket.Conn
	RemoveCh   chan *websocket.Conn
	MsgCh      chan types.Message
	QuitCh     chan struct{}
	Database   *sql.DB
	mutex      sync.Mutex
}

func NewServer(listenAddr string) *Server {
	return &Server{
		ListenAddr: listenAddr,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Clients:  make(map[*websocket.Conn]bool),
		AddCh:    make(chan *websocket.Conn),
		RemoveCh: make(chan *websocket.Conn),
		QuitCh:   make(chan struct{}),
		MsgCh:    make(chan types.Message),
		mutex:    sync.Mutex{},
		Database: db.CreateDb("../database/database.sql"),
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/ws", s.handleWS)

	go s.broadcastLoop()
	go s.listenForExit()

	fmt.Println("Server listening on: ", s.ListenAddr)

	return http.ListenAndServe(s.ListenAddr, nil)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}

	s.AddCh <- conn

	go s.readLoop(conn)
}

func (s *Server) readLoop(conn *websocket.Conn) {
	defer func() {
		s.RemoveCh <- conn
		conn.Close()
	}()

	for {
		var msg types.Envelope
		if err := conn.ReadJSON(&msg); err != nil {
			fmt.Println("read json error: ", err)
			s.RemoveCh <- conn
			return
		}

		switch msg.Type {
		case "init":
			fmt.Println("Initial messsage from " + utils.NormalizeAddr(conn.RemoteAddr().String()))
		case "chat":
			var m types.Message
			if err := json.Unmarshal(msg.Payload, &m); err != nil {
				log.Println("Unmarshal error: ", err)
				return
			}
			s.MsgCh <- types.Message{
				From:    utils.NormalizeAddr(conn.RemoteAddr().String()),
				Payload: bytes.TrimSpace(m.Payload),
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
			s.mutex.Lock()
			s.Clients[conn] = true
			s.mutex.Unlock()
			fmt.Println("New client connected")

		case conn := <-s.RemoveCh:
			s.mutex.Lock()
			if _, ok := s.Clients[conn]; ok {
				delete(s.Clients, conn)
				conn.Close()
				fmt.Println("Client disconnected")
			}
			s.mutex.Unlock()

		case msg := <-s.MsgCh:
			fmt.Printf("Message from %s: %s\n", msg.From, msg.Payload)
			s.mutex.Lock()
			for conn := range s.Clients {
				if utils.NormalizeAddr(conn.RemoteAddr().String()) == msg.From {
					continue
				}
				if err := conn.WriteJSON(msg); err != nil {
					fmt.Println("write error:", err)
					s.RemoveCh <- conn
				}
			}
			s.mutex.Unlock()

		case <-s.QuitCh:
			return
		}
	}
}

func (s *Server) listenForExit() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if scanner.Text() == "exit" {
			fmt.Println("Shutting down server...")

			close(s.QuitCh)

			s.mutex.Lock()
			env := types.NewEnvelope("exit", nil)
			for conn := range s.Clients {
				if err := conn.WriteJSON(env); err != nil {
					fmt.Println("write error:", err)
					continue
				}
				conn.Close()
			}
			s.Clients = make(map[*websocket.Conn]bool)
			s.mutex.Unlock()

			os.Exit(0)
		}
	}
}
