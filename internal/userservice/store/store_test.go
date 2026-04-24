package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Ow1Dev/felter/internal/testutil"
)

func TestPostgresStore(t *testing.T) {
	pool := dbtest.StartPostgres(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store := NewPostgresStore(pool)

	t.Run("create and list", func(t *testing.T) {
		// Use unique values to avoid collision across repeated test runs.
		email := fmt.Sprintf("alice-%d@test.com", time.Now().UnixNano())
		username := fmt.Sprintf("alice-%d", time.Now().UnixNano())

		created, err := store.CreateUser(ctx, email, username, "Alice")
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.ID == 0 {
			t.Fatal("expected non-zero id")
		}

		list, err := store.ListUsers(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		var found bool
		for _, u := range list {
			if u.ID == created.ID {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("created user not in list")
		}
	})
}
