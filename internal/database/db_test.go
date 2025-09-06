package db

import (
	"database/sql"
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
}

func TestGetUsername(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		db      *sql.DB
		user    *types.User
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := GetUsername(tt.db, tt.user)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetUsername() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetUsername() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}
