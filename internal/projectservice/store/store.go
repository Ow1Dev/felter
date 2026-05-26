// Package store defines persistence operations for projects.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/Ow1Dev/felter/internal/projectservice/api"
)

// ErrProjectNotFound is returned when a project is not found.
var ErrProjectNotFound = fmt.Errorf("project not found")

// ErrInvalidProjectName is returned when a project name cannot produce a valid slug.
var ErrInvalidProjectName = fmt.Errorf("invalid project name")

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

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	b.Grow(len(s))

	var prevHyphen bool
	for _, r := range s {
		if r == '_' {
			r = '-'
		}
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevHyphen = false
		case r == '-':
			if !prevHyphen {
				b.WriteRune(r)
				prevHyphen = true
			}
		default:
			if !prevHyphen {
				b.WriteRune('-')
				prevHyphen = true
			}
		}
	}

	return strings.Trim(b.String(), "-")
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
	baseSlug := slugify(name)
	if baseSlug == "" {
		return nil, ErrInvalidProjectName
	}

	slug, err := s.makeUniqueSlug(ctx, baseSlug)
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
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			// Unique violation on slug — retry with a suffix.
			return s.createProjectWithSuffix(ctx, name, description, baseSlug, 1)
		}
		return nil, fmt.Errorf("create project: %w", err)
	}
	if desc.Valid {
		p.Description = &desc.String
	}
	return &p, nil
}

func (s *PostgresStore) createProjectWithSuffix(ctx context.Context, name, description, base string, start int) (*api.Project, error) {
	for i := start; i <= 1000; i++ {
		slug := fmt.Sprintf("%s-%d", base, i)
		const q = `
			INSERT INTO projects (name, description, slug)
			VALUES ($1, $2, $3)
			RETURNING id, name, description, slug, created_at, updated_at
		`
		var p api.Project
		var desc sql.NullString
		err := s.db.QueryRowContext(ctx, q, name, description, slug).Scan(
			&p.Id, &p.Name, &desc, &p.Slug, &p.CreatedAt, &p.UpdatedAt,
		)
		if err == nil {
			if desc.Valid {
				p.Description = &desc.String
			}
			return &p, nil
		}
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			continue
		}
		return nil, fmt.Errorf("create project: %w", err)
	}
	return nil, fmt.Errorf("could not generate unique slug for %q", base)
}

// ListProjects returns all projects ordered by creation time.
func (s *PostgresStore) ListProjects(ctx context.Context) ([]api.Project, error) {
	const q = `SELECT id, name, description, slug, created_at, updated_at FROM projects ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]api.Project, 0)
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
