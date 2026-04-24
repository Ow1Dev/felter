// Package dbtest provides test helpers for database-backed tests.
package dbtest

import (
	"context"
	"database/sql"
	"io"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Ow1Dev/felter/internal/db"
	"github.com/Ow1Dev/felter/internal/migrate"
)

// StartPostgres starts a Postgres container, applies all discovered migrations,
// and returns a *sql.DB connected to it. The container is terminated when the
// test finishes.
func StartPostgres(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("felter"),
		postgres.WithUsername("felter"),
		postgres.WithPassword("felter"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Skipf("start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	pool, err := db.Open(connStr)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	t.Cleanup(func() {
		if err := pool.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	})

	// Apply all discovered migrations exactly as cmd/migrate does.
	services, err := migrate.DiscoverServices("internal")
	if err != nil {
		t.Fatalf("discover migrations: %v", err)
	}
	if len(services) == 0 {
		t.Log("no migration directories found")
		return pool
	}

	results := migrate.Up(pool, services, io.Discard, io.Discard)
	for _, res := range results {
		if res.Err != nil {
			t.Fatalf("migrate %s: %v", res.Service, res.Err)
		}
		t.Logf("[OK] %s: applied %d migration(s)", res.Service, res.Applied)
	}

	return pool
}
