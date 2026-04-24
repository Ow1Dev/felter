// Package main applies all pending SQL migrations across internal/**/migrations.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Ow1Dev/felter/internal/db"
	"github.com/Ow1Dev/felter/internal/migrate"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(
	_ context.Context,
	_ []string,
	getenv func(string) string,
	_ io.Reader,
	stdout, stderr io.Writer,
) error {
	dsn := getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://felter:felter@localhost:5432/felter?sslmode=disable"
	}

	pool, err := db.Open(dsn)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			_, _ = fmt.Fprintf(stderr, "db close: %v\n", err)
		}
	}()

	services, err := migrate.DiscoverServices("internal")
	if err != nil {
		return fmt.Errorf("discover services: %w", err)
	}

	if len(services) == 0 {
		_, _ = fmt.Fprintln(stdout, "no migration directories found")
		return nil
	}

	results := migrate.Up(pool, services, stdout, stderr)
	var failed int
	for _, res := range results {
		if res.Err != nil {
			failed++
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d services failed", failed, len(results))
	}
	_, _ = fmt.Fprintf(stdout, "done: checked %d service(s)\n", len(services))
	return nil
}
