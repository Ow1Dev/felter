// Package store defines persistence operations for schemas and schema fields.
package store

import (
	"context"
	"database/sql"
	"fmt"
)

// ErrSchemaNotFound is returned when a schema is not found.
var ErrSchemaNotFound = fmt.Errorf("schema not found")

// ErrSchemaAlreadyExists is returned when a schema key conflicts.
var ErrSchemaAlreadyExists = fmt.Errorf("schema already exists")

// ErrFieldAlreadyExists is returned when a field key conflicts.
var ErrFieldAlreadyExists = fmt.Errorf("field already exists")

// ErrFieldNotFound is returned when a field is not found.
var ErrFieldNotFound = fmt.Errorf("field not found")

// ErrRecordNotFound is returned when a record is not found.
var ErrRecordNotFound = fmt.Errorf("record not found")

// PostgresStore implements Store with lib/pq.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore returns a Store backed by Postgres.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) resolveSchemaID(ctx context.Context, projectSlug, schemaKey string) (string, error) {
	const q = `SELECT id FROM schemas WHERE project_slug = $1 AND key = $2`
	var id string
	err := s.db.QueryRowContext(ctx, q, projectSlug, schemaKey).Scan(&id)
	if err == sql.ErrNoRows {
		return "", ErrSchemaNotFound
	}
	if err != nil {
		return "", fmt.Errorf("resolve schema id: %w", err)
	}
	return id, nil
}
