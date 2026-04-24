// Package migrate provides a shared migration engine that discovers per-service
// migration directories and applies them using golang-migrate with isolated
// tracking tables.
package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gmigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	// Register the file:// source driver for golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Service holds the name and absolute path to one service's migrations.
type Service struct {
	Name string
	Path string
}

// Result reports the outcome of migrating a single service.
type Result struct {
	Service string
	Applied int
	Err     error
}

// DiscoverServices walks rootDir and returns every */migrations/ directory
// that contains at least one .sql file.
func DiscoverServices(rootDir string) ([]Service, error) {
	var services []Service

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
		serviceDir := filepath.Dir(path)
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		services = append(services, Service{
			Name: filepath.Base(serviceDir),
			Path: absPath,
		})
		return filepath.SkipDir
	})
	if err != nil {
		return nil, err
	}

	return services, nil
}

// Up applies all pending up migrations for every discovered service.
// It writes progress to stdout/stderr and returns a slice of per-service results.
func Up(pool *sql.DB, services []Service, stdout, stderr io.Writer) []Result {
	results := make([]Result, 0, len(services))
	for _, svc := range services {
		applied, err := migrateServiceUp(pool, svc, stderr)
		results = append(results, Result{
			Service: svc.Name,
			Applied: applied,
			Err:     err,
		})
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "[FAIL] %s: %v\n", svc.Name, err)
			continue
		}
		if applied == 0 {
			_, _ = fmt.Fprintf(stdout, "[OK]   %s: already up to date\n", svc.Name)
		} else {
			_, _ = fmt.Fprintf(stdout, "[OK]   %s: applied %d migration(s)\n", svc.Name, applied)
		}
	}
	return results
}

func migrateServiceUp(pool *sql.DB, svc Service, stderr io.Writer) (int, error) {
	driver, err := postgres.WithInstance(pool, &postgres.Config{
		MigrationsTable: svc.Name + "_migrations",
	})
	if err != nil {
		return 0, fmt.Errorf("postgres driver: %w", err)
	}

	sourceURL := "file://" + svc.Path
	m, err := gmigrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return 0, fmt.Errorf("migrate instance: %w", err)
	}
	defer func() {
		_, dbErr := m.Close()
		if dbErr != nil {
			_, _ = fmt.Fprintf(stderr, "migrate close (%s): %v\n", svc.Name, dbErr)
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, gmigrate.ErrNoChange) {
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
