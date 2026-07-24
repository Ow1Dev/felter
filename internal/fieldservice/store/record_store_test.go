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

	// Record that omits the filtered field to verify missing fields don't abort the query.
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"due_date": "2026-07-16T00:00:00Z",
	}, false, 1)
	if err != nil {
		t.Fatalf("create missing priority: %v", err)
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
		{Key: "priority", Type: api.Int},
	})

	rec, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "Fix bug",
		"priority": float64(1),
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
	if len(records[0].Values) != 1 {
		t.Fatalf("expected 1 value after field deletion, got %v", records[0].Values)
	}
	if records[0].Values["priority"] != int64(1) {
		t.Fatalf("expected priority to remain, got %v", records[0].Values)
	}

	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM field_values WHERE record_id = $1 AND field_key = $2`, rec.RecordId.String(), "title").Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 raw rows for deleted field in field_values, got %d", count)
	}
}

func TestQueryRecords_NeMissingField(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("nems-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
		{Key: "priority", Type: api.Int},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "A",
	}, false, 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Record is missing "priority", so ne on priority should match all records.
	filter := &FilterNode{Op: OpNe, Field: "priority", Value: float64(99)}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for ne on missing field, got %d", len(records))
	}
}

func TestQueryRecords_NeDifferentValue(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("nedv-%d", time.Now().UnixNano())
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

	filter := &FilterNode{Op: OpNe, Field: "priority", Value: float64(1)}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for ne different value, got %d", len(records))
	}
	if records[0].Values["priority"] != int64(2) {
		t.Fatalf("expected priority 2, got %v", records[0].Values["priority"])
	}
}

func TestQueryRecords_LikeSpecialChars(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("lsc-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "100% complete",
	}, false, 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "under_score",
	}, false, 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	filter := &FilterNode{Op: OpLike, Field: "title", Value: "%"}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for literal %% match, got %d", len(records))
	}
	if records[0].Values["title"] != "100% complete" {
		t.Fatalf("unexpected record: %v", records[0].Values["title"])
	}

	filter2 := &FilterNode{Op: OpLike, Field: "title", Value: "_"}
	records2, err := s.QueryRecords(ctx, projectSlug, "task", filter2)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records2) != 1 {
		t.Fatalf("expected 1 record for literal _ match, got %d", len(records2))
	}
	if records2[0].Values["title"] != "under_score" {
		t.Fatalf("unexpected record: %v", records2[0].Values["title"])
	}
}

func TestQueryRecords_UnknownField(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("uf-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "A",
	}, false, 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// eq on unknown field returns no records.
	filter := &FilterNode{Op: OpEq, Field: "unknown", Value: "x"}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records for unknown field eq, got %d", len(records))
	}

	// ne on unknown field returns all records.
	filter2 := &FilterNode{Op: OpNe, Field: "unknown", Value: "x"}
	records2, err := s.QueryRecords(ctx, projectSlug, "task", filter2)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records2) != 1 {
		t.Fatalf("expected 1 record for unknown field ne, got %d", len(records2))
	}
}

func TestQueryRecords_BooleanFilter(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("bf-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "active", Type: api.Boolean},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"active": true,
	}, false, 1)
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"active": false,
	}, false, 1)
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	filter := &FilterNode{Op: OpEq, Field: "active", Value: true}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for bool eq, got %d", len(records))
	}
	if records[0].Values["active"] != true {
		t.Fatalf("expected active true, got %v", records[0].Values["active"])
	}

	filter2 := &FilterNode{Op: OpNe, Field: "active", Value: true}
	records2, err := s.QueryRecords(ctx, projectSlug, "task", filter2)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records2) != 1 {
		t.Fatalf("expected 1 record for bool ne, got %d", len(records2))
	}
	if records2[0].Values["active"] != false {
		t.Fatalf("expected active false, got %v", records2[0].Values["active"])
	}
}

func TestQueryRecords_FloatFilter(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("ff-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "rating", Type: api.Float},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"rating": float64(1.5),
	}, false, 1)
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"rating": float64(3.7),
	}, false, 1)
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	filter := &FilterNode{Op: OpGt, Field: "rating", Value: float64(2.0)}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for float gt, got %d", len(records))
	}
	if records[0].Values["rating"] != float64(3.7) {
		t.Fatalf("expected rating 3.7, got %v", records[0].Values["rating"])
	}

	filter2 := &FilterNode{Op: OpEq, Field: "rating", Value: float64(1.5)}
	records2, err := s.QueryRecords(ctx, projectSlug, "task", filter2)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(records2) != 1 {
		t.Fatalf("expected 1 record for float eq, got %d", len(records2))
	}
	if records2[0].Values["rating"] != float64(1.5) {
		t.Fatalf("expected rating 1.5, got %v", records2[0].Values["rating"])
	}
}

func TestQueryRecords_EmptyAndOr(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("eaor-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "A",
	}, false, 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Empty AND should match all records.
	filter := &FilterNode{Op: OpAnd, Conditions: []FilterNode{}}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query empty and: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for empty AND, got %d", len(records))
	}

	// Empty OR should match no records.
	filter2 := &FilterNode{Op: OpOr, Conditions: []FilterNode{}}
	records2, err := s.QueryRecords(ctx, projectSlug, "task", filter2)
	if err != nil {
		t.Fatalf("query empty or: %v", err)
	}
	if len(records2) != 0 {
		t.Fatalf("expected 0 records for empty OR, got %d", len(records2))
	}
}

func TestQueryRecords_NestedAndOr(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("nao-%d", time.Now().UnixNano())
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
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "Refactor",
		"priority": float64(3),
	}, false, 1)
	if err != nil {
		t.Fatalf("create third: %v", err)
	}

	// AND( OR(title = "Fix bug", title = "Add feature"), priority > 1 )
	// Should return only "Add feature" because "Fix bug" has priority 1.
	filter := &FilterNode{
		Op: OpAnd,
		Conditions: []FilterNode{
			{
				Op: OpOr,
				Conditions: []FilterNode{
					{Op: OpEq, Field: "title", Value: "Fix bug"},
					{Op: OpEq, Field: "title", Value: "Add feature"},
				},
			},
			{Op: OpGt, Field: "priority", Value: float64(1)},
		},
	}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query nested and/or: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Values["title"] != "Add feature" {
		t.Fatalf("unexpected record title: %v", records[0].Values["title"])
	}
}

func TestQueryRecords_NestedAndWithNeEq(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("nane-%d", time.Now().UnixNano())
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
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "Refactor",
		"priority": float64(3),
	}, false, 1)
	if err != nil {
		t.Fatalf("create third: %v", err)
	}

	// AND( Ne(title, "Fix bug"), Eq(priority, 2) )
	// Should return only "Add feature".
	filter := &FilterNode{
		Op: OpAnd,
		Conditions: []FilterNode{
			{Op: OpNe, Field: "title", Value: "Fix bug"},
			{Op: OpEq, Field: "priority", Value: float64(2)},
		},
	}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query and with ne/eq: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Values["title"] != "Add feature" {
		t.Fatalf("unexpected record title: %v", records[0].Values["title"])
	}
}

func TestQueryRecords_DateOffsetComparison(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx := context.Background()

	projectSlug := fmt.Sprintf("doc-%d", time.Now().UnixNano())
	setupSchemaWithFields(t, s, projectSlug, "task", []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "due_date", Type: api.Date},
	})

	_, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"due_date": "2026-07-15T00:00:00Z",
	}, false, 1)
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	_, err = s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"due_date": "2026-07-20T00:00:00+00:00",
	}, false, 1)
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	// Query with a different offset but same instant; gt should match the second record chronologically.
	filter := &FilterNode{Op: OpGt, Field: "due_date", Value: "2026-07-18T00:00:00Z"}
	records, err := s.QueryRecords(ctx, projectSlug, "task", filter)
	if err != nil {
		t.Fatalf("query date gt: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	// The record with 2026-07-20 should match.
	expectedDate := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	gotDate, ok := records[0].Values["due_date"].(time.Time)
	if !ok || !gotDate.Equal(expectedDate) {
		t.Fatalf("unexpected record due_date: %v", records[0].Values["due_date"])
	}
}
