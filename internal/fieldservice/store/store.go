// Package store defines persistence operations for schemas and schema fields.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
	"github.com/Ow1Dev/felter/internal/fieldservice/api/fieldvalue"
	"github.com/Ow1Dev/felter/internal/fieldservice/validation"
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

// FilterOp represents a filter operation.
type FilterOp string

// Filter operators for record queries.
const (
	OpEq   FilterOp = "eq"   // Equal
	OpNe   FilterOp = "ne"   // Not equal
	OpGt   FilterOp = "gt"   // Greater than
	OpGte  FilterOp = "gte"  // Greater than or equal
	OpLt   FilterOp = "lt"   // Less than
	OpLte  FilterOp = "lte"  // Less than or equal
	OpLike FilterOp = "like" // Substring match
	OpAnd  FilterOp = "and"  // Logical AND
	OpOr   FilterOp = "or"   // Logical OR
)

// FilterNode is a recursive AST node for in-memory record filtering.
type FilterNode struct {
	Op         FilterOp
	Field      string
	Value      any
	Conditions []FilterNode
}

// Store defines persistence operations for schemas and fields.
type Store interface {
	CreateSchema(ctx context.Context, projectSlug, key, name string, createdBy int64) (*api.Schema, error)
	ListSchemasByProject(ctx context.Context, projectSlug string) ([]api.Schema, error)
	GetSchemaWithFields(ctx context.Context, projectSlug, schemaKey string) (*api.Schema, []api.SchemaField, error)
	DeleteSchema(ctx context.Context, projectSlug, schemaKey string) error
	CreateSchemaField(ctx context.Context, projectSlug, schemaKey, fieldKey string, fieldType api.FieldType, createdBy int64) (*api.SchemaField, error)
	DeleteSchemaField(ctx context.Context, projectSlug, schemaKey, fieldKey string) error
	MutateRecord(ctx context.Context, projectSlug, schemaKey string, recordID *string, values map[string]any, deleteRecord bool, createdBy int64) (*fieldvalue.FieldRecord, error)
	QueryRecords(ctx context.Context, projectSlug, schemaKey string, filter *FilterNode) ([]fieldvalue.FieldRecord, error)
	GetSchemaDefinition(ctx context.Context, projectSlug, schemaKey string) (*fieldvalue.SchemaDefinition, error)
}

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

	const q = `
		SELECT record_id, field_key, value, created_at, updated_at
		FROM field_values
		WHERE schema_id = $1
	`
	rows, err := s.db.QueryContext(ctx, q, schemaID)
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

	if filter == nil {
		return out, nil
	}

	if err := validateFilter(filter); err != nil {
		return nil, err
	}

	filtered := make([]fieldvalue.FieldRecord, 0, len(out))
	for _, rec := range out {
		match, err := evaluateFilter(filter, rec)
		if err != nil {
			return nil, err
		}
		if match {
			filtered = append(filtered, rec)
		}
	}
	return filtered, nil
}

func validateFilter(filter *FilterNode) error {
	if filter == nil {
		return nil
	}
	switch filter.Op {
	case OpEq, OpNe, OpGt, OpGte, OpLt, OpLte, OpLike:
		return nil
	case OpAnd, OpOr:
		for i := range filter.Conditions {
			if err := validateFilter(&filter.Conditions[i]); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown filter operator: %s", filter.Op)
	}
}

func evaluateFilter(filter *FilterNode, rec fieldvalue.FieldRecord) (bool, error) {
	switch filter.Op {
	case OpAnd:
		for _, c := range filter.Conditions {
			match, err := evaluateFilter(&c, rec)
			if err != nil {
				return false, err
			}
			if !match {
				return false, nil
			}
		}
		return true, nil
	case OpOr:
		for _, c := range filter.Conditions {
			match, err := evaluateFilter(&c, rec)
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
		return false, nil
	case OpEq, OpNe, OpGt, OpGte, OpLt, OpLte, OpLike:
		return evaluateLeafFilter(filter, rec)
	default:
		return false, fmt.Errorf("unknown filter operator: %s", filter.Op)
	}
}

func evaluateLeafFilter(filter *FilterNode, rec fieldvalue.FieldRecord) (bool, error) {
	var left any
	if filter.Field == "record_id" {
		left = rec.RecordId.String()
	} else {
		var ok bool
		left, ok = rec.Values[filter.Field]
		if !ok {
			// Field not present on record.
			switch filter.Op {
			case OpEq:
				return false, nil
			case OpNe:
				return true, nil
			default:
				return false, fmt.Errorf("field %q not found in record", filter.Field)
			}
		}
	}

	// Coerce filter value when the record value is time.Time and filter value is a string.
	right := filter.Value
	if lt, ok := left.(time.Time); ok {
		if rs, ok := right.(string); ok {
			if pt, err := time.Parse(time.RFC3339, rs); err == nil {
				right = pt
			} else {
				return false, fmt.Errorf("cannot parse %q as RFC3339 for field %q", rs, filter.Field)
			}
		} else {
			return false, fmt.Errorf("cannot compare time.Time with %T for field %q", right, filter.Field)
		}
		_ = lt
	}

	switch filter.Op {
	case OpEq:
		return compareEqual(left, right), nil
	case OpNe:
		return !compareEqual(left, right), nil
	case OpGt:
		return compareGreater(left, right)
	case OpGte:
		gt, err := compareGreater(left, right)
		if err != nil {
			return false, err
		}
		return gt || compareEqual(left, right), nil
	case OpLt:
		gt, err := compareGreater(left, right)
		if err != nil {
			return false, err
		}
		return !gt && !compareEqual(left, right), nil
	case OpLte:
		gt, err := compareGreater(left, right)
		if err != nil {
			return false, err
		}
		return !gt, nil
	case OpLike:
		ls, lok := left.(string)
		rs, rok := right.(string)
		if !lok || !rok {
			return false, fmt.Errorf("like operator requires string values")
		}
		return strings.Contains(strings.ToLower(ls), strings.ToLower(rs)), nil
	default:
		return false, fmt.Errorf("unknown filter operator: %s", filter.Op)
	}
}

func compareEqual(a, b any) bool {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case int64:
		switch bv := b.(type) {
		case int64:
			return av == bv
		case float64:
			return float64(av) == bv
		default:
			return false
		}
	case float64:
		switch bv := b.(type) {
		case float64:
			return av == bv
		case int64:
			return av == float64(bv)
		default:
			return false
		}
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case time.Time:
		bv, ok := b.(time.Time)
		return ok && av.Equal(bv)
	default:
		return false
	}
}

func compareGreater(a, b any) (bool, error) {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		if !ok {
			return false, fmt.Errorf("cannot compare string with %T", b)
		}
		return av > bv, nil
	case int64:
		switch bv := b.(type) {
		case int64:
			return av > bv, nil
		case float64:
			return float64(av) > bv, nil
		default:
			return false, fmt.Errorf("cannot compare int64 with %T", b)
		}
	case float64:
		switch bv := b.(type) {
		case float64:
			return av > bv, nil
		case int64:
			return av > float64(bv), nil
		default:
			return false, fmt.Errorf("cannot compare float64 with %T", b)
		}
	case time.Time:
		bv, ok := b.(time.Time)
		if !ok {
			return false, fmt.Errorf("cannot compare time.Time with %T", b)
		}
		return av.After(bv), nil
	default:
		return false, fmt.Errorf("unsupported type for comparison: %T", a)
	}
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
