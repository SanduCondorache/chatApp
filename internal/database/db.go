package db

import (
	"database/sql"
	"log"

	"github.com/SanduCondorache/chatApp/internal/types"
	_ "github.com/mattn/go-sqlite3"
)

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

func InsertUser(db *sql.DB, user *types.User) error {
	_, err := db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", user.Username, user.Email, user.Password)
	return err
}

func GetUserId(db *sql.DB, user *types.User) (int, error) {
	rows, err := db.Query("SELECT id FROM users WHERE username = ?", user.Username)

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

func GetUserByUsername(db *sql.DB, username string) (*types.User, error) {
	u := &types.User{}
	query := "SELECT username, email, password FROM users WHERE username = ?"

	err := db.QueryRow(query, username).Scan(&u.Username, &u.Email, &u.Password)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func GetUsername(db *sql.DB, username string) (bool, error) {
	rows, err := db.Query("SELECT 1 FROM users WHERE username = ?", username)

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

func insertMessage(db *sql.DB, sender_id, recipient_id int, content string) error {
	_, err := db.Exec("INSERT INTO messages (sender_id, recipient_id, content) VALUES (?, ?, ?)", sender_id, recipient_id, content)
	return err
}

// func getUserMessagesBy(db *sql.DB, sender_id int) ([]string, error) {
// 	rows, err := db.Query("SELECT DISTINCT u.username FROM m messages JOIN users u ON u.id = m.recipient_id WHERE m.sender_id = ?", sender_id)
// 	if err != nil {
// 		return []string{}, err
// 	}
//
// 	defer rows.Close()
// }

func CheckTablesExists(db *sql.DB) (bool, error) {
	var hasTable bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM sqlite_master
			WHERE type='table'
			  AND name NOT LIKE 'sqlite_%'
			LIMIT 1
		);
	`).Scan(&hasTable)

	if err != nil {
		return false, err
	}

	return hasTable, nil

}

func UserExists(db *sql.DB, username string) (bool, error) {
	var sw bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`

	err := db.QueryRow(query, username).Scan(&sw)
	if err != nil {
		return false, err
	}

	return sw, nil
}
