package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
	"github.com/Ow1Dev/felter/internal/fieldservice/validation"
	"github.com/Ow1Dev/felter/internal/testutil"
)

func setupSchemaWithFields(t *testing.T, s *PostgresStore, projectSlug, schemaKey string, fields []struct {
	Key  string
	Type api.FieldType
},
) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.CreateSchema(ctx, projectSlug, schemaKey, schemaKey, 1); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	for _, f := range fields {
		if _, err := s.CreateSchemaField(ctx, projectSlug, schemaKey, f.Key, f.Type, 1); err != nil {
			t.Fatalf("create field %q: %v", f.Key, err)
		}
	}
}

func TestMutateRecord_Create(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("create-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
		{Key: "priority", Type: api.Int},
		{Key: "due_date", Type: api.Date},
	})

	rec, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "Fix bug",
		"priority": float64(1),
		"due_date": "2026-07-15T00:00:00Z",
	}, false, 1)
	if err != nil {
		t.Fatalf("mutate create: %v", err)
	}
	if rec.RecordId.String() == "" {
		t.Fatal("expected non-empty record_id")
	}
	if rec.SchemaKey != "task" || rec.ProjectSlug != projectSlug {
		t.Fatalf("unexpected record meta: %+v", rec)
	}
	if rec.Values["title"] != "Fix bug" {
		t.Fatalf("title mismatch: %v", rec.Values["title"])
	}
	if rec.Values["priority"] != int64(1) {
		t.Fatalf("priority mismatch: %v", rec.Values["priority"])
	}
	if rec.CreatedAt.IsZero() || rec.UpdatedAt.IsZero() {
		t.Fatal("expected timestamps set")
	}
}

func TestMutateRecord_Update(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("update-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
		{Key: "priority", Type: api.Int},
		{Key: "due_date", Type: api.Date},
	})

	rec, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "Fix bug",
		"priority": float64(1),
		"due_date": "2026-07-15T00:00:00Z",
	}, false, 1)
	if err != nil {
		t.Fatalf("mutate create: %v", err)
	}

	recordID := rec.RecordId.String()
	updated, err := s.MutateRecord(ctx, projectSlug, "task", &recordID, map[string]any{
		"title": "Updated title",
	}, false, 1)
	if err != nil {
		t.Fatalf("mutate update: %v", err)
	}
	if updated.Values["title"] != "Updated title" {
		t.Fatalf("title mismatch: %v", updated.Values["title"])
	}
	if updated.Values["priority"] != int64(1) {
		t.Fatalf("priority should remain: %v", updated.Values["priority"])
	}
	if updated.UpdatedAt.Before(rec.CreatedAt) {
		t.Fatal("updated_at should not be before created_at")
	}
}

func TestMutateRecord_DeleteField(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("delfield-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
		{Key: "priority", Type: api.Int},
	})

	rec, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "Fix bug",
		"priority": float64(1),
	}, false, 1)
	if err != nil {
		t.Fatalf("mutate create: %v", err)
	}

	recordID := rec.RecordId.String()
	updated, err := s.MutateRecord(ctx, projectSlug, "task", &recordID, map[string]any{
		"title": nil,
	}, false, 1)
	if err != nil {
		t.Fatalf("mutate update: %v", err)
	}
	if _, ok := updated.Values["title"]; ok {
		t.Fatal("expected title to be deleted")
	}
	if updated.Values["priority"] != int64(1) {
		t.Fatalf("priority should remain: %v", updated.Values["priority"])
	}
}

func TestMutateRecord_DeleteRecord(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("delrec-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
	})

	rec, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "Fix bug",
	}, false, 1)
	if err != nil {
		t.Fatalf("mutate create: %v", err)
	}

	recordID := rec.RecordId.String()
	deleted, err := s.MutateRecord(ctx, projectSlug, "task", &recordID, nil, true, 1)
	if err != nil {
		t.Fatalf("mutate delete: %v", err)
	}
	if deleted != nil {
		t.Fatal("expected nil record after delete")
	}

	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM field_values WHERE record_id = $1`, recordID).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows, got %d", count)
	}
}

func TestMutateRecord_SchemaNotFound(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("snf-%d", time.Now().UnixNano())
	_, err := s.MutateRecord(ctx, projectSlug, "nonexistent", nil, map[string]any{
		"title": "x",
	}, false, 1)
	if err != ErrSchemaNotFound {
		t.Fatalf("expected ErrSchemaNotFound, got %v", err)
	}
}

func TestMutateRecord_InvalidFieldKey(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("ifk-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"unknown": "x",
	}, false, 1)
	if err == nil {
		t.Fatal("expected error for unknown field key")
	}
	if _, ok := err.(validation.FieldError); !ok {
		t.Fatalf("expected FieldError, got %T", err)
	}
}

func TestMutateRecord_TypeMismatch(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("tm-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "priority", Type: api.Int},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"priority": "not a number",
	}, false, 1)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
	fe, ok := err.(validation.FieldError)
	if !ok {
		t.Fatalf("expected FieldError, got %T", err)
	}
	if fe.Field != "priority" {
		t.Fatalf("field = %q, want %q", fe.Field, "priority")
	}
}

func TestQueryRecords_ListAll(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("listall-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
		{Key: "priority", Type: api.Int},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "A",
		"priority": float64(1),
	}, false, 1)
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "B",
		"priority": float64(2),
	}, false, 1)
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	records, err := s.QueryRecords(ctx, projectSlug, "task", nil)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}

func TestQueryRecords_ByRecordID(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("qid-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
	})

	rec, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "A",
	}, false, 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	filter := &FilterNode{Op: OpEq, Field: "record_id", Value: rec.RecordId.String()}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	filter2 := &FilterNode{Op: OpEq, Field: "record_id", Value: "00000000-0000-0000-0000-000000000000"}
	records2, err := s.QueryRecords(ctx, projectSlug, "task", filter2)
	if err != nil {
		t.Fatalf("query nonexistent: %v", err)
	}
	if len(records2) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records2))
	}
}

func TestQueryRecords_ByFieldValue(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("qfv-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "priority", Type: api.Int},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"priority": float64(1),
	}, false, 1)
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"priority": float64(2),
	}, false, 1)
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	filter := &FilterNode{Op: OpEq, Field: "priority", Value: float64(1)}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}

func TestQueryRecords_RangeFilter(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("qrf-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "priority", Type: api.Int},
		{Key: "due_date", Type: api.Date},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"priority": float64(1),
		"due_date": "2026-07-10T00:00:00Z",
	}, false, 1)
	if err != nil {
		t.Fatalf("create low: %v", err)
	}
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"priority": float64(5),
		"due_date": "2026-07-20T00:00:00Z",
	}, false, 1)
	if err != nil {
		t.Fatalf("create high: %v", err)
	}

	gtFilter := &FilterNode{Op: OpGt, Field: "priority", Value: float64(3)}
	records, err := s.QueryRecords(ctx, projectSlug, "task", gtFilter)
	if err != nil {
		t.Fatalf("query gt: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for gt, got %d", len(records))
	}

	ltFilter := &FilterNode{Op: OpLt, Field: "due_date", Value: "2026-07-15T00:00:00Z"}
	records2, err := s.QueryRecords(ctx, projectSlug, "task", ltFilter)
	if err != nil {
		t.Fatalf("query lt date: %v", err)
	}
	if len(records2) != 1 {
		t.Fatalf("expected 1 record for lt date, got %d", len(records2))
	}
}

func TestQueryRecords_LikeFilter(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("qlf-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "Fix the nasty bug",
	}, false, 1)
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "Add feature",
	}, false, 1)
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	filter := &FilterNode{Op: OpLike, Field: "title", Value: "BUG"}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query like: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Values["title"] != "Fix the nasty bug" {
		t.Fatalf("unexpected record: %v", records[0].Values["title"])
	}
}

func TestQueryRecords_AndOrCombinators(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("qaoc-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
		{Key: "priority", Type: api.Int},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "Fix bug",
		"priority": float64(1),
	}, false, 1)
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "Add feature",
		"priority": float64(2),
	}, false, 1)
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	andFilter := &FilterNode{
		Op: OpAnd,
		Conditions: []FilterNode{
			{Op: OpEq, Field: "title", Value: "Fix bug"},
			{Op: OpEq, Field: "priority", Value: float64(1)},
		},
	}
	records, err := s.QueryRecords(ctx, projectSlug, "task", andFilter)
	if err != nil {
		t.Fatalf("query and: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for and, got %d", len(records))
	}

	orFilter := &FilterNode{
		Op: OpOr,
		Conditions: []FilterNode{
			{Op: OpEq, Field: "priority", Value: float64(99)},
			{Op: OpEq, Field: "title", Value: "Add feature"},
		},
	}
	records2, err := s.QueryRecords(ctx, projectSlug, "task", orFilter)
	if err != nil {
		t.Fatalf("query or: %v", err)
	}
	if len(records2) != 1 {
		t.Fatalf("expected 1 record for or, got %d", len(records2))
	}
}

func TestGetSchemaDefinition(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("gsd-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
		{Key: "priority", Type: api.Int},
	})

	def, err := s.GetSchemaDefinition(ctx, projectSlug, "task")
	if err != nil {
		t.Fatalf("get definition: %v", err)
	}
	if def.SchemaKey != "task" || def.ProjectSlug != projectSlug {
		t.Fatalf("unexpected definition meta: %+v", def)
	}
	if len(def.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(def.Fields))
	}
	found := false
	for _, f := range def.Fields {
		if f.Key == "priority" && f.Type == "int" && !f.Required {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected priority field definition, got %+v", def.Fields)
	}
}

func TestSchemaChangeResilience(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("scr-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
	})

	rec, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "Fix bug",
	}, false, 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := s.DeleteSchemaField(ctx, projectSlug, "task", "title"); err != nil {
		t.Fatalf("delete field: %v", err)
	}

	records, err := s.QueryRecords(ctx, projectSlug, "task", nil)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record after field deletion, got %d", len(records))
	}
	if len(records[0].Values) != 0 {
		t.Fatalf("expected 0 values after field deletion, got %v", records[0].Values)
	}

	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM field_values WHERE record_id = $1`, rec.RecordId.String()).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 raw row in field_values, got %d", count)
	}
}
