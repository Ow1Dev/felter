// Package main applies all pending SQL migrations across internal/**/migrations.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/Ow1Dev/felter/internal/db"
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

	services, err := discoverServices("internal")
	if err != nil {
		return fmt.Errorf("discover services: %w", err)
	}

	if len(services) == 0 {
		_, _ = fmt.Fprintln(stdout, "no migration directories found")
		return nil
	}

	var failed int
	for _, svc := range services {
		applied, err := migrateService(pool, svc, stderr)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "[FAIL] %s: %v\n", svc.name, err)
			failed++
			continue
		}
		if applied == 0 {
			_, _ = fmt.Fprintf(stdout, "[OK]   %s: already up to date\n", svc.name)
		} else {
			_, _ = fmt.Fprintf(stdout, "[OK]   %s: applied %d migration(s)\n", svc.name, applied)
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d services failed", failed, len(services))
	}
	_, _ = fmt.Fprintf(stdout, "done: checked %d service(s)\n", len(services))
	return nil
}

// serviceMigrations holds the path to one service's migrations directory.
type serviceMigrations struct {
	name string // service name (e.g., "userservice")
	path string // absolute path to the migrations directory
}

// discoverServices walks rootDir and returns every directory matching
// */migrations/ that contains at least one .sql file.
func discoverServices(rootDir string) ([]serviceMigrations, error) {
	var services []serviceMigrations

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if filepath.Base(path) != "migrations" {
			return nil
		}
		// Check the directory has at least one .sql file.
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		hasSQL := false
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
				hasSQL = true
				break
			}
		}
		if !hasSQL {
			return nil
		}
		// Service name is the parent of the migrations directory.
		serviceDir := filepath.Dir(path)
		serviceName := filepath.Base(serviceDir)
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		services = append(services, serviceMigrations{
			name: serviceName,
			path: absPath,
		})
		return filepath.SkipDir // don't recurse into migrations/
	})
	if err != nil {
		return nil, err
	}

	return services, nil
}

// migrateService runs pending migrations for a single service.
// It returns the number of migrations that were applied.
func migrateService(pool *sql.DB, svc serviceMigrations, stderr io.Writer) (int, error) {
	// Use a per-service migrations table to keep services isolated.
	driver, err := postgres.WithInstance(pool, &postgres.Config{
		MigrationsTable: svc.name + "_migrations",
	})
	if err != nil {
		return 0, fmt.Errorf("postgres driver: %w", err)
	}

	sourceURL := "file://" + svc.path
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return 0, fmt.Errorf("migrate instance: %w", err)
	}
	defer func() {
		_, dbErr := m.Close()
		if dbErr != nil {
			_, _ = fmt.Fprintf(stderr, "migrate close (%s): %v\n", svc.name, dbErr)
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return 0, nil
		}
		return 0, fmt.Errorf("up: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		return 0, fmt.Errorf("version: %w", err)
	}
	if dirty {
		return int(version), fmt.Errorf("migration %d is dirty", version)
	}
	return int(version), nil
}
