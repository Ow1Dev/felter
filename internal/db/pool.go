// Package db provides a reusable PostgreSQL connection pool.
package db

import (
	"database/sql"
	"fmt"

	// import lib/pq driver for database/sql
	_ "github.com/lib/pq"
)

// Open opens and pings a *sql.DB for the given Postgres DSN.
func Open(dsn string) (*sql.DB, error) {
	pool, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := pool.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}
