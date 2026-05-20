// Package store defines persistence operations for projects.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/Ow1Dev/felter/internal/projectservice/api"
)

// ErrProjectNotFound is returned when a project is not found.
var ErrProjectNotFound = fmt.Errorf("project not found")

// Store defines persistence operations for projects.
type Store interface {
	CreateProject(ctx context.Context, name, description string) (*api.Project, error)
	ListProjects(ctx context.Context) ([]api.Project, error)
	GetProjectBySlug(ctx context.Context, slug string) (*api.Project, error)
}

// PostgresStore implements Store with lib/pq.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore returns a Store backed by Postgres.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

var (
	slugInvalidChars = regexp.MustCompile(`[^a-z0-9-]+`)
	slugHyphenRepeat = regexp.MustCompile(`-+`)
)

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, "_", "-")
	s = slugInvalidChars.ReplaceAllString(s, "-")
	s = slugHyphenRepeat.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func (s *PostgresStore) makeUniqueSlug(ctx context.Context, base string) (string, error) {
	slug := base
	for i := 1; i <= 1000; i++ {
		var exists bool
		err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE slug = $1)`, slug).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("check slug exists: %w", err)
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
	return "", fmt.Errorf("could not generate unique slug for %q", base)
}

// CreateProject inserts a new project and returns it.
func (s *PostgresStore) CreateProject(ctx context.Context, name, description string) (*api.Project, error) {
	slug, err := s.makeUniqueSlug(ctx, slugify(name))
	if err != nil {
		return nil, err
	}

	const q = `
		INSERT INTO projects (name, description, slug)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, slug, created_at, updated_at
	`
	var p api.Project
	var desc sql.NullString
	if err := s.db.QueryRowContext(ctx, q, name, description, slug).Scan(
		&p.Id, &p.Name, &desc, &p.Slug, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	if desc.Valid {
		p.Description = &desc.String
	}
	return &p, nil
}

// ListProjects returns all projects ordered by creation time.
func (s *PostgresStore) ListProjects(ctx context.Context) ([]api.Project, error) {
	const q = `SELECT id, name, description, slug, created_at, updated_at FROM projects ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []api.Project
	for rows.Next() {
		var p api.Project
		var desc sql.NullString
		if err := rows.Scan(&p.Id, &p.Name, &desc, &p.Slug, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		if desc.Valid {
			p.Description = &desc.String
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// GetProjectBySlug returns a project by its slug.
func (s *PostgresStore) GetProjectBySlug(ctx context.Context, slug string) (*api.Project, error) {
	const q = `SELECT id, name, description, slug, created_at, updated_at FROM projects WHERE slug = $1`
	var p api.Project
	var desc sql.NullString
	err := s.db.QueryRowContext(ctx, q, slug).Scan(&p.Id, &p.Name, &desc, &p.Slug, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get project by slug: %w", err)
	}
	if desc.Valid {
		p.Description = &desc.String
	}
	return &p, nil
}
