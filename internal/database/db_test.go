package db

import (
	"fmt"
	"testing"

	"github.com/SanduCondorache/chatApp/internal/types"
)

func TestDatabase(t *testing.T) {
	db := CreateDb("./database.sql")

	defer db.Close()

	user := types.NewUser("loh", "loh@gmail.com", "123455")
	err := InsertUser(db, user)
	if err != nil {
		t.Fatalf("Failed to insert the user")
	}

	id, err := GetUserId(db, user)
	if err != nil {
		t.Fatalf("Failed to query on db")
	}

	if id != 1 {
		t.Fatalf("Incorect query got %d", id)
	}

	err = InsertUser(db, user)
	if err != nil {
		fmt.Println(err)
		t.Fatalf(err.Error())
	}

}
