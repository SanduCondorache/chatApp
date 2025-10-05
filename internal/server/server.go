package server

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
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
	Database   *dab.Store
	mutex      sync.Mutex
}

func CreateServer(listenAddr string, db *dab.Store) *Server {
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
		log.Println("DB file not found, creating and initializing DB...")
		db := dab.NewStore(dbPath)
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
		return CreateServer(listenAddr, dab.NewStore(dbPath))
	}

	return CreateServer(listenAddr, dab.NewStore(db))
}

func (s *Server) Start() error {
	http.HandleFunc("/ws", s.handleWS)

	go s.broadcastLoop()
	go s.listenForCommands()

	log.Println("Server listening on: ", s.ListenAddr)

	return http.ListenAndServe(s.ListenAddr, nil)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	s.AddCh <- conn

	go s.readLoop(conn)
}

func sendMessageFromServer(t types.MessageType, payload string, conn *websocket.Conn) error {
	msg := types.NewMessage(payload)
	data, err := msg.ToEnvelopePayload()
	if err != nil {
		return err
	}

	env := types.NewEnvelope(t, data)
	return conn.WriteJSON(&env)
}

func (s *Server) loginUser(msg types.Envelope, conn *websocket.Conn) error {
	user, err := types.ReadUser(msg, conn)
	if err != nil {
		return err
	}

	exists, err := s.Database.UserExists(user.Username)
	if err != nil {
		sendMessageFromServer(types.Error, "user_not_found", conn)
		return err
	}

	hasedPassword, err := s.Database.GetPassword(user)

	sw := utils.ComparePasswords(hasedPassword, user.Password)

	if !sw {
		sendMessageFromServer(types.Error, "incorrect_password", conn)
		return err
	}

	log.Println(user.Username, "has logged in")

	s.mutex.Lock()
	s.Clients[conn] = user
	s.ClientsRev[user.Username] = conn
	s.mutex.Unlock()

	if exists {
		sendMessageFromServer(types.Ok, "ok", conn)
		return err
	}

	return types.ErrorUserNotFound
}

func (s *Server) registerUser(msg types.Envelope, conn *websocket.Conn) error {
	user, err := types.ReadUser(msg, conn)
	if err != nil {
		return err
	}

	err = s.Database.InsertUser(user)
	log.Println(user)
	if err != nil {
		var errr sqlite3.Error
		if errors.As(err, &errr) && errr.Code == sqlite3.ErrConstraint {
			sendMessageFromServer(types.Error, "username_taken", conn)
			log.Println("Username is already used")
			return nil
		}
		return err
	}

	sendMessageFromServer(types.Ok, "ok", conn)
	s.mutex.Lock()
	s.Clients[conn] = user
	s.ClientsRev[user.Username] = conn
	s.mutex.Unlock()

	return nil

}

func (s *Server) registerOrLoginUser(msg types.Envelope, conn *websocket.Conn) error {
	if msg.Type == types.Login {
		return s.loginUser(msg, conn)
	}

	return s.registerUser(msg, conn)
}

func (s *Server) handleChatMessages(msg types.Envelope, conn *websocket.Conn) error {
	var m types.ChatMessage
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		return err
	}
	log.Println(m)

	err := s.Database.InsertMessage(&m)
	if err != nil {
		return err
	}

	sendMessageFromServer(types.MsgSent, "ok", conn)

	return nil
}

func (s *Server) findUser(msg types.Envelope, conn *websocket.Conn) error {
	var m types.Message
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		return err
	}

	exists, err := s.Database.GetUsername(string(m.Payload))

	if err != nil {
		return err
	}

	if !exists {
		sendMessageFromServer(types.Error, "user_not_found", conn)
		return nil
	}

	sendMessageFromServer(types.Ok, "ok", conn)

	return nil
}

func (s *Server) checkOnlineUsers(msg types.Envelope, conn *websocket.Conn) error {
	var m types.Message
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		return err
	}

	var mp map[string][]string

	if err := json.Unmarshal(m.Payload, &mp); err != nil {
		return err
	}

	temp := s.getUsersOnline(mp)

	data, err := json.Marshal(temp)
	if err != nil {
		return err
	}

	sendMessageFromServer(types.GetConn, string(data), conn)

	return nil
}

func (s *Server) getMessages(msg types.Envelope, conn *websocket.Conn) error {
	var m types.Message
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		return err
	}

	var mp map[string]string

	if err := json.Unmarshal(m.Payload, &mp); err != nil {
		return err
	}

	messages, err := s.Database.GetUserMessagesBy(mp["user1"], mp["user2"])
	if err != nil {
		return err
	}

	sendMessageFromServer(types.GetMsg, messages, conn)

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
			log.Println("read json error: ", err)
			s.RemoveCh <- conn
			return
		}

		switch msg.Type {
		case types.Login, types.Register:
			log.Println("Initial messsage from " + utils.NormalizeAddr(conn.RemoteAddr().String()))
			if err := s.registerOrLoginUser(msg, conn); err != nil {
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
		case types.GetConn:
			if err := s.checkOnlineUsers(msg, conn); err != nil {
				log.Println("reading message err: ", err)
				return
			}
		case types.GetMsg:
			if err := s.getMessages(msg, conn); err != nil {
				log.Println("reading message err: ", err)
				return
			}
		default:
			log.Println("unknown message type ", msg.Type)
		}

	}
}

func (s *Server) broadcastLoop() {
	for {
		select {
		case conn := <-s.AddCh:
			log.Println("New client connected", utils.NormalizeAddr(conn.RemoteAddr().String()))
		case conn := <-s.RemoveCh:
			s.mutex.Lock()
			if _, ok := s.Clients[conn]; ok {
				u := s.Clients[conn]
				delete(s.ClientsRev, u.Username)
				delete(s.Clients, conn)
				conn.Close()
				log.Println("Client disconnected")
			}
			s.mutex.Unlock()

		case msg := <-s.MsgCh:
			s.mutex.Lock()
			log.Println(msg)
			s.mutex.Unlock()

		case <-s.QuitCh:
			return
		}
	}
}

func (s *Server) listenForCommands() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if scanner.Text() == "exit" {
			log.Println("Shutting down server...")

			close(s.QuitCh)

			s.mutex.Lock()
			env := types.NewEnvelope(types.Exit, nil)
			for conn := range s.Clients {
				if err := conn.WriteJSON(env); err != nil {
					log.Println("write error:", err)
					continue
				}
				conn.Close()
			}
			s.Clients = make(map[*websocket.Conn]*types.User)
			s.mutex.Unlock()

			os.Exit(0)
		} else if scanner.Text() == "get" {
			s.mutex.Lock()
			log.Println("ClientsRev:", len(s.ClientsRev))
			for user := range s.ClientsRev {
				log.Println(user)
			}
			log.Println("Clients:", len(s.Clients))
			for conn, user := range s.Clients {
				log.Printf("%s -> %v\n", utils.NormalizeAddr(conn.RemoteAddr().String()), user.Username)
			}
			s.mutex.Unlock()
		}
	}
}

func (s *Server) getUserConn(username string) (*websocket.Conn, error) {
	u, err := s.Database.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	return s.ClientsRev[u.Username], nil
}

func (s *Server) getUsersOnline(m map[string][]string) map[string]bool {
	res := make(map[string]bool)

	users := m["users"]

	for _, u := range users {
		if _, ok := s.ClientsRev[u]; ok {
			res[u] = true
		} else {
			res[u] = false
		}
	}

	return res
}
