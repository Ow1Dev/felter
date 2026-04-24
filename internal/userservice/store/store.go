// Package store defines persistence operations for users.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// User is the domain model.
type User struct {
	ID          int64
	Email       string
	Username    string
	DisplayName *string
	CreatedAt   time.Time
}

// Store defines persistence operations for users.
type Store interface {
	CreateUser(ctx context.Context, email, username, displayName string) (*User, error)
	ListUsers(ctx context.Context) ([]*User, error)
}

// PostgresStore implements Store with lib/pq.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore returns a Store backed by Postgres.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// CreateUser inserts a new user and returns it.
func (s *PostgresStore) CreateUser(ctx context.Context, email, username, displayName string) (*User, error) {
	const q = `
		INSERT INTO users (email, username, display_name)
		VALUES ($1, $2, $3)
		RETURNING id, email, username, display_name, created_at
	`
	var u User
	var dn sql.NullString
	if err := s.db.QueryRowContext(ctx, q, email, username, displayName).Scan(
		&u.ID, &u.Email, &u.Username, &dn, &u.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	if dn.Valid {
		u.DisplayName = &dn.String
	}
	return &u, nil
}

// ListUsers returns all users ordered by creation time.
func (s *PostgresStore) ListUsers(ctx context.Context) ([]*User, error) {
	const q = `SELECT id, email, username, display_name, created_at FROM users ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*User
	for rows.Next() {
		var u User
		var dn sql.NullString
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &dn, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		if dn.Valid {
			u.DisplayName = &dn.String
		}
		out = append(out, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}
