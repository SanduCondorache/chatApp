package server

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/SanduCondorache/chatApp/internal/config"
	dab "github.com/SanduCondorache/chatApp/internal/database"
	"github.com/SanduCondorache/chatApp/internal/types"
	"github.com/SanduCondorache/chatApp/utils"
	"github.com/gorilla/websocket"
	"github.com/mattn/go-sqlite3"
)

type Server struct {
	ListenAddr string
	Upgrader   websocket.Upgrader
	Clients    map[*websocket.Conn]*types.User
	ClientsRev map[string]*websocket.Conn
	AddCh      chan *websocket.Conn
	RemoveCh   chan *websocket.Conn
	MsgCh      chan types.ChatMessage
	QuitCh     chan struct{}
	Database   *sql.DB
	mutex      sync.Mutex
}

func CreateServer(listenAddr string, db *sql.DB) *Server {
	return &Server{
		ListenAddr: listenAddr,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Clients:    make(map[*websocket.Conn]*types.User),
		AddCh:      make(chan *websocket.Conn),
		RemoveCh:   make(chan *websocket.Conn),
		QuitCh:     make(chan struct{}),
		MsgCh:      make(chan types.ChatMessage),
		mutex:      sync.Mutex{},
		Database:   db,
		ClientsRev: map[string]*websocket.Conn{},
	}
}

func NewServer(listenAddr string) *Server {
	dbPath := config.Envs.DBPath

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("DB file not found, creating and initializing DB...")
		db := dab.CreateDb(dbPath)
		return CreateServer(listenAddr, db)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Println("Opening database error: ", err)
		return nil
	}

	has, err := dab.CheckTablesExists(db)

	if err != nil {
		log.Println("query error: ", err)
	}

	if !has {
		return CreateServer(listenAddr, dab.CreateDb(dbPath))
	}

	return CreateServer(listenAddr, db)
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

func sendMessageFromServer(t types.MessageType, payload string, conn *websocket.Conn) {
	msg := types.NewMessage(payload)
	data, _ := msg.ToEnvelopePayload()
	env := types.NewEnvelope(t, data)
	_ = conn.WriteJSON(&env)
}

func (s *Server) readUser(msg types.Envelope, conn *websocket.Conn) error {
	if msg.Type == types.Login {
		u := &types.User{}
		if err := json.Unmarshal(msg.Payload, u); err != nil {
			return err
		}
		exists, err := dab.UserExists(s.Database, u.Username)
		if err != nil {
			return err
		}

		fmt.Println("Inserted user", u, err)
		s.mutex.Lock()
		s.Clients[conn] = u
		s.ClientsRev[u.Username] = conn
		s.mutex.Unlock()

		if exists {
			// eror := types.NewMessage("ok")
			// data, _ := eror.ToEnvelopePayload()
			// env := types.NewEnvelope("ok", data)
			// _ = conn.WriteJSON(&env)
			sendMessageFromServer(types.Ok, "ok", conn)
			return nil
		}

		return types.ErrorUserNotFound

	}
	var u types.User
	if err := json.Unmarshal(msg.Payload, &u); err != nil {
		return err
	}

	err := dab.InsertUser(s.Database, &u)
	fmt.Println(u, err)
	if err != nil {
		var errr sqlite3.Error
		if errors.As(err, &errr) && errr.Code == sqlite3.ErrConstraint {
			// eror := types.NewMessage("username_taken")
			// data, _ := eror.ToEnvelopePayload()
			// env := types.NewEnvelope("error", data)
			// _ = conn.WriteJSON(&env)
			sendMessageFromServer(types.Error, "username_taken", conn)

			log.Println("Username is already used")
			return nil
		}
		return err
	}

	// eror := types.NewMessage("ok")
	// data, _ := eror.ToEnvelopePayload()
	// env := types.NewEnvelope("ok", data)
	// _ = conn.WriteJSON(&env)
	sendMessageFromServer(types.Ok, "ok", conn)
	s.mutex.Lock()
	s.Clients[conn] = &u
	s.ClientsRev[u.Username] = conn
	s.mutex.Unlock()

	return nil
}

func (s *Server) handleChatMessages(msg types.Envelope, conn *websocket.Conn) error {
	var m types.ChatMessage
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		return err
	}

	fmt.Println(m)
	receiver := m.Recv

	receiver_conn, err := s.getUserConn(receiver)
	fmt.Println("Is nil?", receiver_conn == nil)
	fmt.Println(utils.NormalizeAddr(receiver_conn.RemoteAddr().String()))
	if err != nil {
		return err
	}

	receiver_msg := types.NewEnvelope(types.MsgRecv, msg.Payload)
	err = receiver_conn.WriteJSON(receiver_msg)
	if err != nil {
		return err
	}

	// TODO: Insert the message into db

	// eror := types.NewMessage("ok")
	// data, _ := eror.ToEnvelopePayload()
	// env := types.NewEnvelope(types.Ok, data)
	// _ = conn.WriteJSON(&env)

	sendMessageFromServer(types.MsgSent, "ok", conn)

	return nil
}

func (s *Server) findUser(msg types.Envelope, conn *websocket.Conn) error {
	var m types.Message
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		return err
	}

	exists, err := dab.GetUsername(s.Database, string(m.Payload))
	fmt.Println(exists)

	if err != nil {
		return err
	}

	if !exists {
		eror := types.NewMessage("user_not_found")
		data, _ := eror.ToEnvelopePayload()
		env := types.NewEnvelope("error", data)
		_ = conn.WriteJSON(&env)

		return nil
	}

	// eror := types.NewMessage("ok")
	// data, _ := eror.ToEnvelopePayload()
	// env := types.NewEnvelope("ok", data)
	// _ = conn.WriteJSON(&env)
	sendMessageFromServer(types.Ok, "ok", conn)

	return nil
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
		case types.Login:
			fmt.Println("Initial messsage from " + utils.NormalizeAddr(conn.RemoteAddr().String()))
			if err := s.readUser(msg, conn); err != nil {
				log.Println("reading user err: ", err)
				return
			}
		case types.Chat:
			if err := s.handleChatMessages(msg, conn); err != nil {
				log.Println("reading message err: ", err)
				return
			}
		case types.Find:
			if err := s.findUser(msg, conn); err != nil {
				log.Println("reading message err: ", err)
				return
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
			fmt.Println("New client connected", utils.NormalizeAddr(conn.RemoteAddr().String()))
		case conn := <-s.RemoveCh:
			s.mutex.Lock()
			if _, ok := s.Clients[conn]; ok {
				u := s.Clients[conn]
				delete(s.ClientsRev, u.Username)
				delete(s.Clients, conn)
				conn.Close()
				fmt.Println("Client disconnected")
			}
			s.mutex.Unlock()

		case msg := <-s.MsgCh:
			s.mutex.Lock()
			fmt.Println(msg)
			// for conn := range s.Clients {
			// 	if utils.NormalizeAddr(conn.RemoteAddr().String()) == msg.From {
			// 		continue
			// 	}
			// 	if err := conn.WriteJSON(msg); err != nil {
			// 		fmt.Println("write error:", err)
			// 		s.RemoveCh <- conn
			// 	}
			// }
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
			env := types.NewEnvelope(types.Exit, nil)
			for conn := range s.Clients {
				if err := conn.WriteJSON(env); err != nil {
					fmt.Println("write error:", err)
					continue
				}
				conn.Close()
			}
			s.Clients = make(map[*websocket.Conn]*types.User)
			s.mutex.Unlock()

			os.Exit(0)
		} else if scanner.Text() == "get" {
			s.mutex.Lock()
			fmt.Println("ClientsRev:", len(s.ClientsRev))
			for user := range s.ClientsRev {
				fmt.Println(user)
			}
			fmt.Println("Clients:", len(s.Clients))
			for conn, user := range s.Clients {
				fmt.Printf("%s -> %v\n", utils.NormalizeAddr(conn.RemoteAddr().String()), user.Username)
			}
			s.mutex.Unlock()
		}
	}
}

func (s *Server) getUserConn(username string) (*websocket.Conn, error) {
	s.mutex.Lock()
	for client := range s.ClientsRev {
		fmt.Println(client)
	}
	for conns := range s.Clients {
		fmt.Println(utils.NormalizeAddr(conns.RemoteAddr().String()))
	}
	s.mutex.Unlock()
	u, err := dab.GetUserByUsername(s.Database, username)
	fmt.Println("User from db", u, len(s.Clients))
	if err != nil {
		return nil, err
	}
	return s.ClientsRev[u.Username], nil
}
