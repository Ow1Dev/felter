package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
)

// SchemaStore defines persistence operations for schemas and schema fields.
type SchemaStore interface {
	CreateSchema(ctx context.Context, projectSlug, key, name string, createdBy int64) (*api.Schema, error)
	ListSchemasByProject(ctx context.Context, projectSlug string) ([]api.Schema, error)
	GetSchemaWithFields(ctx context.Context, projectSlug, schemaKey string) (*api.Schema, []api.SchemaField, error)
	DeleteSchema(ctx context.Context, projectSlug, schemaKey string) error
	CreateSchemaField(ctx context.Context, projectSlug, schemaKey, fieldKey string, fieldType api.FieldType, createdBy int64) (*api.SchemaField, error)
	DeleteSchemaField(ctx context.Context, projectSlug, schemaKey, fieldKey string) error
}

// CreateSchema inserts a new schema and returns it.
func (s *PostgresStore) CreateSchema(ctx context.Context, projectSlug, key, name string, createdBy int64) (*api.Schema, error) {
	const q = `
		INSERT INTO schemas (project_slug, key, name, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING project_slug, key, name, created_by, created_at, updated_at
	`
	var sch api.Schema
	if err := s.db.QueryRowContext(ctx, q, projectSlug, key, name, createdBy).Scan(
		&sch.ProjectSlug, &sch.Key, &sch.Name, &sch.CreatedBy, &sch.CreatedAt, &sch.UpdatedAt,
	); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return nil, ErrSchemaAlreadyExists
		}
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &sch, nil
}

// ListSchemasByProject returns all schemas for a project ordered by creation time.
func (s *PostgresStore) ListSchemasByProject(ctx context.Context, projectSlug string) ([]api.Schema, error) {
	const q = `
		SELECT project_slug, key, name, created_by, created_at, updated_at
		FROM schemas
		WHERE project_slug = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, q, projectSlug)
	if err != nil {
		return nil, fmt.Errorf("list schemas: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]api.Schema, 0)
	for rows.Next() {
		var sch api.Schema
		if err := rows.Scan(&sch.ProjectSlug, &sch.Key, &sch.Name, &sch.CreatedBy, &sch.CreatedAt, &sch.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan schema: %w", err)
		}
		out = append(out, sch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// GetSchemaWithFields returns a schema and its fields.
func (s *PostgresStore) GetSchemaWithFields(ctx context.Context, projectSlug, schemaKey string) (*api.Schema, []api.SchemaField, error) {
	const qSchema = `
		SELECT project_slug, key, name, created_by, created_at, updated_at
		FROM schemas
		WHERE project_slug = $1 AND key = $2
	`
	var sch api.Schema
	err := s.db.QueryRowContext(ctx, qSchema, projectSlug, schemaKey).Scan(
		&sch.ProjectSlug, &sch.Key, &sch.Name, &sch.CreatedBy, &sch.CreatedAt, &sch.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil, ErrSchemaNotFound
	}
	if err != nil {
		return nil, nil, fmt.Errorf("get schema: %w", err)
	}

	const qFields = `
		SELECT f.key, f.type, f.created_by, f.created_at, f.updated_at
		FROM schema_fields f
		JOIN schemas s ON s.id = f.schema_id
		WHERE s.project_slug = $1 AND s.key = $2
		ORDER BY f.created_at ASC
	`
	rows, err := s.db.QueryContext(ctx, qFields, projectSlug, schemaKey)
	if err != nil {
		return nil, nil, fmt.Errorf("list fields: %w", err)
	}
	defer func() { _ = rows.Close() }()

	fields := make([]api.SchemaField, 0)
	for rows.Next() {
		var f api.SchemaField
		if err := rows.Scan(&f.Key, &f.Type, &f.CreatedBy, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, nil, fmt.Errorf("scan field: %w", err)
		}
		fields = append(fields, f)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("field rows err: %w", err)
	}
	return &sch, fields, nil
}

// DeleteSchema removes a schema and its fields.
func (s *PostgresStore) DeleteSchema(ctx context.Context, projectSlug, schemaKey string) error {
	const q = `DELETE FROM schemas WHERE project_slug = $1 AND key = $2`
	res, err := s.db.ExecContext(ctx, q, projectSlug, schemaKey)
	if err != nil {
		return fmt.Errorf("delete schema: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrSchemaNotFound
	}
	return nil
}

// CreateSchemaField inserts a new field into a schema.
func (s *PostgresStore) CreateSchemaField(ctx context.Context, projectSlug, schemaKey, fieldKey string, fieldType api.FieldType, createdBy int64) (*api.SchemaField, error) {
	schemaID, err := s.resolveSchemaID(ctx, projectSlug, schemaKey)
	if err != nil {
		return nil, err
	}

	const q = `
		INSERT INTO schema_fields (schema_id, key, type, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING key, type, created_by, created_at, updated_at
	`
	var f api.SchemaField
	if err := s.db.QueryRowContext(ctx, q, schemaID, fieldKey, fieldType, createdBy).Scan(
		&f.Key, &f.Type, &f.CreatedBy, &f.CreatedAt, &f.UpdatedAt,
	); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return nil, ErrFieldAlreadyExists
		}
		return nil, fmt.Errorf("create field: %w", err)
	}
	return &f, nil
}

// DeleteSchemaField removes a field from a schema.
func (s *PostgresStore) DeleteSchemaField(ctx context.Context, projectSlug, schemaKey, fieldKey string) error {
	schemaID, err := s.resolveSchemaID(ctx, projectSlug, schemaKey)
	if err != nil {
		return err
	}

	const q = `DELETE FROM schema_fields WHERE schema_id = $1 AND key = $2`
	res, err := s.db.ExecContext(ctx, q, schemaID, fieldKey)
	if err != nil {
		return fmt.Errorf("delete field: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrFieldNotFound
	}
	return nil
}
