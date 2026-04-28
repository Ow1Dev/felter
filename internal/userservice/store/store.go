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

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = fmt.Errorf("user not found")

// Store defines persistence operations for users.
type Store interface {
	CreateUser(ctx context.Context, email, username, displayName string) (*User, error)
	ListUsers(ctx context.Context) ([]*User, error)
	GetUser(ctx context.Context, id int64) (*User, error)
	GetUserFromProvider(ctx context.Context, provider, providerID string) (*User, error)
	CreateUserFromProvider(ctx context.Context, provider, providerID, email, username string) (*User, error)
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

func (s *PostgresStore) GetUser(ctx context.Context, id int64) (*User, error) {
	const q = `SELECT id, email, username, display_name, created_at FROM users WHERE id = $1`
	var u User
	var dn sql.NullString
	err := s.db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Email, &u.Username, &dn, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if dn.Valid {
		u.DisplayName = &dn.String
	}
	return &u, nil
}

func (s *PostgresStore) GetUserFromProvider(ctx context.Context, provider, providerID string) (*User, error) {
	const q = `
		SELECT u.id, u.email, u.username, u.display_name, u.created_at
		FROM users u
		JOIN user_identities ui ON ui.user_id = u.id
		WHERE ui.provider = $1 AND ui.provider_id = $2
	`
	var u User
	var dn sql.NullString
	err := s.db.QueryRowContext(ctx, q, provider, providerID).Scan(
		&u.ID, &u.Email, &u.Username, &dn, &u.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user from provider: %w", err)
	}
	if dn.Valid {
		u.DisplayName = &dn.String
	}
	return &u, nil
}

func (s *PostgresStore) CreateUserFromProvider(ctx context.Context, provider, providerID, email, username string) (*User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var u User
	var dn sql.NullString
	const userQ = `
		INSERT INTO users (email, username)
		VALUES ($1, $2)
		RETURNING id, email, username, display_name, created_at
	`
	err = tx.QueryRowContext(ctx, userQ, email, username).Scan(
		&u.ID, &u.Email, &u.Username, &dn, &u.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	const identityQ = `
		INSERT INTO user_identities (user_id, provider, provider_id)
		VALUES ($1, $2, $3)
	`
	_, err = tx.ExecContext(ctx, identityQ, u.ID, provider, providerID)
	if err != nil {
		return nil, fmt.Errorf("create user identity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	if dn.Valid {
		u.DisplayName = &dn.String
	}
	return &u, nil
}
