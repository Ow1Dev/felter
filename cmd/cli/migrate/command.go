// Package migrate implements the "migrate" subcommand for the CLI.
package migrate

import (
	"context"
	"fmt"
	"io"

	"github.com/Ow1Dev/felter/internal/db"
)

// Run executes the migrate subcommand.
func Run(
	_ context.Context,
	_ []string,
	getenv func(string) string,
	_ io.Reader,
	stdout, stderr io.Writer,
) error {
	dsn := getenv("DATABASE_DSN")
	if dsn == "" {
		return fmt.Errorf("DATABASE_DSN is required")
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

	services, err := DiscoverServices("internal")
	if err != nil {
		return fmt.Errorf("discover services: %w", err)
	}

	if len(services) == 0 {
		_, _ = fmt.Fprintln(stdout, "no migration directories found")
		return nil
	}

	results := Up(pool, services, stdout, stderr)
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
