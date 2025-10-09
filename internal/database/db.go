package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/SanduCondorache/chatApp/internal/types"
	"github.com/SanduCondorache/chatApp/utils"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func CreateDb(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Println("Opening database error: ", err)
		return nil
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")

	if err != nil {
		log.Println("Exec error: ", err)
		return nil
	}

	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
		email TEXT,
		password TEXT NOT NULL
    );
    CREATE TABLE IF NOT EXISTS messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        sender_id INTEGER NOT NULL,
        recipient_id INTEGER NOT NULL,
        content TEXT NOT NULL,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (sender_id) REFERENCES users(id),
        FOREIGN KEY (recipient_id) REFERENCES users(id)
    );`

	_, err = db.Exec(schema)

	if err != nil {
		log.Println("Exec error: ", err)
		return nil
	}

	return db
}

func NewStore(arg any) *Store {
	switch v := arg.(type) {
	case string:
		return &Store{db: CreateDb(v)}
	case *sql.DB:
		return &Store{db: v}
	default:
		panic("unsupported argument type")
	}
}

func (s *Store) InsertUser(user *types.User) error {
	query := "INSERT INTO users (username, email, password) VALUES (?, ?, ?)"

	pass, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(query, user.Username, user.Email, pass)
	return err
}

func (s *Store) GetUserId(username string) (int, error) {
	query := "SELECT id FROM users WHERE username = ?"

	rows, err := s.db.Query(query, username)

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var id int
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return 0, err
		}
	}

	if err = rows.Err(); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Store) GetUserByUsername(username string) (*types.User, error) {
	u := &types.User{}
	query := "SELECT username, email, password FROM users WHERE username = ?"

	err := s.db.QueryRow(query, username).Scan(&u.Username, &u.Email, &u.Password)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) GetUsername(username string) (bool, error) {
	query := "SELECT 1 FROM users WHERE username = ?"

	rows, err := s.db.Query(query, username)

	if err != nil {
		return false, err
	}

	defer rows.Close()

	var id int
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return false, nil
		}
	}

	if err = rows.Err(); err != nil {
		return false, nil
	}

	exists := id == 1

	return exists, nil
}

func (s *Store) InsertMessage(msg *types.ChatMessage) error {

	sender_id, err := s.GetUserId(msg.Send)
	if err != nil {
		return err
	}

	recipient_id, err := s.GetUserId(msg.Recv)
	if err != nil {
		return err
	}

	query := "INSERT INTO messages (sender_id, recipient_id, content, timestamp) VALUES (?, ?, ?, ?)"

	_, err = s.db.Exec(query, sender_id, recipient_id, msg.Msg, msg.Created_at.String())
	return err
}

func (s *Store) GetUserMessagesBy(sender, recipient string) (string, error) {

	sender_id, err := s.GetUserId(sender)
	if err != nil {
		return "", err
	}

	recipient_id, err := s.GetUserId(recipient)
	if err != nil {
		return "", err
	}

	var messages string
	query := `
		SELECT json_group_array(
			json_object(
				'direction', CASE WHEN sender_id = ? THEN 'sent' ELSE 'received' END,
				'content', content,
				'timestamp', timestamp
			)
		) AS chat_json
		FROM messages
		WHERE (sender_id = ? AND recipient_id = ?)
		OR (sender_id = ? AND recipient_id = ?)
		ORDER BY timestamp;
	`

	err = s.db.QueryRow(query, recipient_id, sender_id, recipient_id, recipient_id, sender_id).Scan(&messages)
	if err != nil {
		return "", err
	}

	return messages, nil
}

func CheckTablesExists(db *sql.DB) (bool, error) {
	var hasTable bool
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM sqlite_master
			WHERE type='table'
			  AND name NOT LIKE 'sqlite_%'
			LIMIT 1
		);
	`

	err := db.QueryRow(query).Scan(&hasTable)

	if err != nil {
		return false, err
	}

	return hasTable, nil

}

func (s *Store) UserExists(username string) (bool, error) {
	var sw bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`

	err := s.db.QueryRow(query, username).Scan(&sw)
	if err != nil {
		return false, err
	}

	return sw, nil
}

func (s *Store) CheckPasswordIsCorrect(user *types.User) (bool, error) {
	var passowrd string
	query := `SELECT password from users WHERE username = ?`

	err := s.db.QueryRow(query, user.Username).Scan(&passowrd)
	if err != nil {
		return false, err
	}

	return passowrd == user.Password, nil
}

func (s *Store) GetPassword(user *types.User) (string, error) {
	var passowrd string
	query := `SELECT password from users WHERE username = ?`

	err := s.db.QueryRow(query, user.Username).Scan(&passowrd)
	if err != nil {
		return "", err
	}

	return passowrd, nil
}

func (s *Store) CheckMessagesBetweenUsersExists(sender string) ([]int, error) {
	query := `
	SELECT DISTINCT
		CASE
			WHEN sender_id = ? THEN recipient_id
			ELSE sender_id
		END AS other_user_id
	FROM messages
	WHERE sender_id = ? OR recipient_id = ?`

	sender_id, err := s.GetUserId(sender)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(query, sender_id, sender_id, sender_id)
	defer rows.Close()

	var chats []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		chats = append(chats, id)
	}

	return chats, nil
}

func (s *Store) GetUsernameById(id int) (string, error) {
	var username string
	query := `SELECT username FROM users WHERE id = ?`

	err := s.db.QueryRow(query, id).Scan(&username)
	fmt.Println(id)
	if err != nil {
		return "", err
	}

	return username, nil
}
