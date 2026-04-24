package store

import (
	"context"
	"testing"
	"time"

	"github.com/Ow1Dev/felter/internal/db"
)

func TestPostgresStore(t *testing.T) {
	dsn := "postgres://felter:felter@localhost:5432/felter?sslmode=disable"
	pool, err := db.Open(dsn)
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store := NewPostgresStore(pool)

	t.Run("create and list", func(t *testing.T) {
		created, err := store.CreateUser(ctx, "alice@test.com", "alice", "Alice")
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
