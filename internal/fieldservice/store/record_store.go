package store

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
	"github.com/Ow1Dev/felter/internal/fieldservice/api/fieldvalue"
	"github.com/Ow1Dev/felter/internal/fieldservice/validation"
)

// RecordStore defines persistence operations for field records.
type RecordStore interface {
	MutateRecord(ctx context.Context, projectSlug, schemaKey string, recordID *string, values map[string]any, deleteRecord bool, createdBy int64) (*fieldvalue.FieldRecord, error)
	QueryRecords(ctx context.Context, projectSlug, schemaKey string, filter *FilterNode) ([]fieldvalue.FieldRecord, error)
	GetSchemaDefinition(ctx context.Context, projectSlug, schemaKey string) (*fieldvalue.SchemaDefinition, error)
}

// schemaFieldMeta holds metadata about a schema field for validation and coercion.
type schemaFieldMeta struct {
	Key  string
	Type api.FieldType
}

// MutateRecord creates, updates, or deletes a record and its field values.
func (s *PostgresStore) MutateRecord(ctx context.Context, projectSlug, schemaKey string, recordID *string, values map[string]any, deleteRecord bool, createdBy int64) (*fieldvalue.FieldRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	schemaID, err := s.resolveSchemaID(ctx, projectSlug, schemaKey)
	if err != nil {
		return nil, err
	}

	// Fetch schema fields into a map.
	fieldsMap, err := s.fetchSchemaFieldsMap(ctx, tx, schemaID)
	if err != nil {
		return nil, err
	}

	if recordID == nil {
		// Create mode.
		newID := uuid.New().String()
		if err := s.validateAndInsertValues(ctx, tx, schemaID, newID, values, fieldsMap, createdBy); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit tx: %w", err)
		}
		return s.assembleRecord(ctx, projectSlug, schemaKey, newID, fieldsMap)
	}

	if deleteRecord {
		// Delete mode.
		const delQ = `DELETE FROM field_values WHERE record_id = $1`
		if _, err := tx.ExecContext(ctx, delQ, *recordID); err != nil {
			return nil, fmt.Errorf("delete record: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit tx: %w", err)
		}
		return nil, nil
	}

	// Update mode.
	const checkQ = `SELECT 1 FROM field_values WHERE record_id = $1 AND schema_id = $2 LIMIT 1`
	var dummy int
	if err := tx.QueryRowContext(ctx, checkQ, *recordID, schemaID).Scan(&dummy); err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	} else if err != nil {
		return nil, fmt.Errorf("check record: %w", err)
	}

	for k, v := range values {
		meta, ok := fieldsMap[k]
		if !ok {
			return nil, validation.FieldError{Field: k, ExpectedType: "known field", Got: "unknown field key"}
		}
		if v == nil {
			const delQ = `DELETE FROM field_values WHERE record_id = $1 AND field_key = $2`
			if _, err := tx.ExecContext(ctx, delQ, *recordID, k); err != nil {
				return nil, fmt.Errorf("delete field value: %w", err)
			}
			continue
		}
		if err := validation.ValidateValue(k, v, meta.Type); err != nil {
			return nil, err
		}
		valStr := valueToString(v)
		const upsertQ = `
			INSERT INTO field_values (schema_id, record_id, field_key, value, created_by)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (record_id, field_key)
			DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
		`
		if _, err := tx.ExecContext(ctx, upsertQ, schemaID, *recordID, k, valStr, createdBy); err != nil {
			return nil, fmt.Errorf("upsert field value: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return s.assembleRecord(ctx, projectSlug, schemaKey, *recordID, fieldsMap)
}

func (s *PostgresStore) fetchSchemaFieldsMap(ctx context.Context, tx *sql.Tx, schemaID string) (map[string]schemaFieldMeta, error) {
	const q = `SELECT key, type FROM schema_fields WHERE schema_id = $1`
	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.QueryContext(ctx, q, schemaID)
	} else {
		rows, err = s.db.QueryContext(ctx, q, schemaID)
	}
	if err != nil {
		return nil, fmt.Errorf("fetch schema fields: %w", err)
	}
	defer func() { _ = rows.Close() }()

	m := make(map[string]schemaFieldMeta)
	for rows.Next() {
		var meta schemaFieldMeta
		if err := rows.Scan(&meta.Key, &meta.Type); err != nil {
			return nil, fmt.Errorf("scan field meta: %w", err)
		}
		m[meta.Key] = meta
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("field meta rows err: %w", err)
	}
	return m, nil
}

func (s *PostgresStore) validateAndInsertValues(ctx context.Context, tx *sql.Tx, schemaID, recordID string, values map[string]any, fieldsMap map[string]schemaFieldMeta, createdBy int64) error {
	for k, v := range values {
		meta, ok := fieldsMap[k]
		if !ok {
			return validation.FieldError{Field: k, ExpectedType: "known field", Got: "unknown field key"}
		}
		if v == nil {
			continue
		}
		if err := validation.ValidateValue(k, v, meta.Type); err != nil {
			return err
		}
		valStr := valueToString(v)
		const q = `
			INSERT INTO field_values (schema_id, record_id, field_key, value, created_by)
			VALUES ($1, $2, $3, $4, $5)
		`
		if _, err := tx.ExecContext(ctx, q, schemaID, recordID, k, valStr, createdBy); err != nil {
			return fmt.Errorf("insert field value: %w", err)
		}
	}
	return nil
}

func valueToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if math.Mod(val, 1.0) == 0 {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (s *PostgresStore) assembleRecord(ctx context.Context, projectSlug, schemaKey, recordID string, fieldsMap map[string]schemaFieldMeta) (*fieldvalue.FieldRecord, error) {
	const q = `
		SELECT field_key, value, created_at, updated_at
		FROM field_values
		WHERE record_id = $1
	`
	rows, err := s.db.QueryContext(ctx, q, recordID)
	if err != nil {
		return nil, fmt.Errorf("assemble record: %w", err)
	}
	defer func() { _ = rows.Close() }()

	values := make(map[string]any)
	var createdAt, updatedAt time.Time
	first := true
	for rows.Next() {
		var fieldKey, valStr string
		var ca, ua time.Time
		if err := rows.Scan(&fieldKey, &valStr, &ca, &ua); err != nil {
			return nil, fmt.Errorf("scan field value: %w", err)
		}
		if first {
			createdAt = ca
			updatedAt = ua
			first = false
		} else {
			if ca.Before(createdAt) {
				createdAt = ca
			}
			if ua.After(updatedAt) {
				updatedAt = ua
			}
		}
		meta, ok := fieldsMap[fieldKey]
		if !ok {
			// Field definition was deleted after value was written; skip it.
			continue
		}
		coerced, err := validation.CoerceValue(valStr, meta.Type)
		if err != nil {
			return nil, fmt.Errorf("coerce %q: %w", fieldKey, err)
		}
		values[fieldKey] = coerced
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("assemble rows err: %w", err)
	}

	// If no rows at all (e.g. empty record), use current time for timestamps.
	if first {
		now := time.Now().UTC()
		createdAt = now
		updatedAt = now
	}

	return &fieldvalue.FieldRecord{
		RecordId:    uuid.MustParse(recordID),
		SchemaKey:   schemaKey,
		ProjectSlug: projectSlug,
		Values:      values,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// QueryRecords returns records for a schema, optionally filtered by an AST.
func (s *PostgresStore) QueryRecords(ctx context.Context, projectSlug, schemaKey string, filter *FilterNode) ([]fieldvalue.FieldRecord, error) {
	schemaID, err := s.resolveSchemaID(ctx, projectSlug, schemaKey)
	if err != nil {
		return nil, err
	}

	fieldsMap, err := s.fetchSchemaFieldsMap(ctx, nil, schemaID)
	if err != nil {
		return nil, err
	}

	q := `
		SELECT record_id, field_key, value, created_at, updated_at
		FROM field_values
		WHERE schema_id = $1
	`
	args := []any{schemaID}

	if filter != nil {
		filterSQL, filterArgs, err := filterToSQL(filter, fieldsMap)
		if err != nil {
			return nil, err
		}
		if filterSQL != "" {
			q += ` AND record_id IN (` + filterSQL + `)`
			args = append(args, filterArgs...)
		}
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query records: %w", err)
	}
	defer func() { _ = rows.Close() }()

	records := make(map[string]*fieldvalue.FieldRecord)
	for rows.Next() {
		var recordID, fieldKey, valStr string
		var ca, ua time.Time
		if err := rows.Scan(&recordID, &fieldKey, &valStr, &ca, &ua); err != nil {
			return nil, fmt.Errorf("scan record row: %w", err)
		}

		rec, ok := records[recordID]
		if !ok {
			rec = &fieldvalue.FieldRecord{
				RecordId:    uuid.MustParse(recordID),
				SchemaKey:   schemaKey,
				ProjectSlug: projectSlug,
				Values:      make(map[string]any),
				CreatedAt:   ca,
				UpdatedAt:   ua,
			}
			records[recordID] = rec
		} else {
			if ca.Before(rec.CreatedAt) {
				rec.CreatedAt = ca
			}
			if ua.After(rec.UpdatedAt) {
				rec.UpdatedAt = ua
			}
		}

		meta, ok := fieldsMap[fieldKey]
		if !ok {
			continue
		}
		coerced, err := validation.CoerceValue(valStr, meta.Type)
		if err != nil {
			return nil, fmt.Errorf("coerce %q: %w", fieldKey, err)
		}
		rec.Values[fieldKey] = coerced
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query rows err: %w", err)
	}

	out := make([]fieldvalue.FieldRecord, 0, len(records))
	for _, rec := range records {
		out = append(out, *rec)
	}

	return out, nil
}

// GetSchemaDefinition returns the schema field definitions for introspection.
func (s *PostgresStore) GetSchemaDefinition(ctx context.Context, projectSlug, schemaKey string) (*fieldvalue.SchemaDefinition, error) {
	schemaID, err := s.resolveSchemaID(ctx, projectSlug, schemaKey)
	if err != nil {
		return nil, err
	}

	const q = `SELECT key, type FROM schema_fields WHERE schema_id = $1 ORDER BY created_at ASC`
	rows, err := s.db.QueryContext(ctx, q, schemaID)
	if err != nil {
		return nil, fmt.Errorf("fetch schema definition: %w", err)
	}
	defer func() { _ = rows.Close() }()

	fields := make([]fieldvalue.FieldDefinition, 0)
	for rows.Next() {
		var fd fieldvalue.FieldDefinition
		var ft api.FieldType
		if err := rows.Scan(&fd.Key, &ft); err != nil {
			return nil, fmt.Errorf("scan field definition: %w", err)
		}
		fd.Type = string(ft)
		fd.Required = false
		fields = append(fields, fd)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("field definition rows err: %w", err)
	}

	return &fieldvalue.SchemaDefinition{
		SchemaKey:   schemaKey,
		ProjectSlug: projectSlug,
		Fields:      fields,
	}, nil
}
