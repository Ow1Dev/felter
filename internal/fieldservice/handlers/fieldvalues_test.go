package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
	"github.com/Ow1Dev/felter/internal/fieldservice/api/fieldvalue"
	"github.com/Ow1Dev/felter/internal/fieldservice/store"
	"github.com/Ow1Dev/felter/internal/testutil"
)

func setupFieldValueStore(t *testing.T) *store.PostgresStore {
	t.Helper()
	pool := dbtest.StartPostgres(t)
	return store.NewPostgresStore(pool)
}

func setupSchemaForFieldValues(t *testing.T, s *store.PostgresStore, projectSlug string) string {
	t.Helper()
	ctx := context.Background()
	if _, err := s.CreateSchema(ctx, projectSlug, "task", "Task", 1); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	for _, f := range []struct {
		Key  string
		Type api.FieldType
	}{
		{Key: "title", Type: api.String},
		{Key: "priority", Type: api.Int},
		{Key: "due_date", Type: api.Date},
	} {
		if _, err := s.CreateSchemaField(ctx, projectSlug, "task", f.Key, f.Type, 1); err != nil {
			t.Fatalf("create field %q: %v", f.Key, err)
		}
	}
	return "task"
}

func TestHandleMutate_Create(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"
	setupSchemaForFieldValues(t, s, projectSlug)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /field-values/mutate", HandleMutateFieldValue(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	values := map[string]any{"title": "Fix login", "priority": float64(1)}
	body, _ := json.Marshal(fieldvalue.MutateRequest{SchemaKey: "task", Values: &values})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/field-values/mutate?project_slug="+projectSlug, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var rec fieldvalue.FieldRecord
	if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if rec.RecordId.String() == "" {
		t.Fatal("expected non-empty record_id")
	}
	if rec.SchemaKey != "task" {
		t.Fatalf("schema_key = %q, want %q", rec.SchemaKey, "task")
	}
	if rec.Values["title"] != "Fix login" {
		t.Fatalf("title = %v, want %v", rec.Values["title"], "Fix login")
	}
}

func TestHandleMutate_Update(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"
	setupSchemaForFieldValues(t, s, projectSlug)

	ctx := context.Background()
	rec, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title":    "Original",
		"priority": float64(1),
	}, false, 1)
	if err != nil {
		t.Fatalf("seed record: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /field-values/mutate", HandleMutateFieldValue(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	rid := rec.RecordId.String()
	values := map[string]any{"title": "Updated"}
	body, _ := json.Marshal(fieldvalue.MutateRequest{SchemaKey: "task", RecordId: &rec.RecordId, Values: &values})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/field-values/mutate?project_slug="+projectSlug, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var updated fieldvalue.FieldRecord
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if updated.Values["title"] != "Updated" {
		t.Fatalf("title = %v, want %v", updated.Values["title"], "Updated")
	}
	if updated.Values["priority"] != float64(1) {
		t.Fatalf("priority = %v, want %v", updated.Values["priority"], float64(1))
	}
	_ = rid
}

func TestHandleMutate_DeleteRecord(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"
	setupSchemaForFieldValues(t, s, projectSlug)

	ctx := context.Background()
	rec, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "To delete",
	}, false, 1)
	if err != nil {
		t.Fatalf("seed record: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /field-values/mutate", HandleMutateFieldValue(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	deleteRecord := true
	body, _ := json.Marshal(fieldvalue.MutateRequest{SchemaKey: "task", RecordId: &rec.RecordId, DeleteRecord: &deleteRecord})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/field-values/mutate?project_slug="+projectSlug, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestHandleMutate_MissingProjectSlug(t *testing.T) {
	s := setupFieldValueStore(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /field-values/mutate", HandleMutateFieldValue(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	body, _ := json.Marshal(fieldvalue.MutateRequest{SchemaKey: "task"})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/field-values/mutate", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleMutate_MissingSchemaKey(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"
	setupSchemaForFieldValues(t, s, projectSlug)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /field-values/mutate", HandleMutateFieldValue(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	body, _ := json.Marshal(fieldvalue.MutateRequest{})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/field-values/mutate?project_slug="+projectSlug, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleMutate_TypeMismatch(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"
	setupSchemaForFieldValues(t, s, projectSlug)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /field-values/mutate", HandleMutateFieldValue(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	values := map[string]any{"priority": "not a number"}
	body, _ := json.Marshal(fieldvalue.MutateRequest{SchemaKey: "task", Values: &values})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/field-values/mutate?project_slug="+projectSlug, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	var er map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if er["error"] != "validation_failed" {
		t.Fatalf("error = %v, want validation_failed", er["error"])
	}
	details, ok := er["details"].([]any)
	if !ok || len(details) == 0 {
		t.Fatalf("expected details array, got %v", er["details"])
	}
}

func TestHandleQuery_List(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"
	setupSchemaForFieldValues(t, s, projectSlug)

	ctx := context.Background()
	if _, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "A", "priority": float64(1),
	}, false, 1); err != nil {
		t.Fatalf("seed 1: %v", err)
	}
	if _, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "B", "priority": float64(2),
	}, false, 1); err != nil {
		t.Fatalf("seed 2: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /field-values/query", HandleQueryFieldValues(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	body, _ := json.Marshal(fieldvalue.QueryRequest{SchemaKey: "task"})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/field-values/query?project_slug="+projectSlug, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var qr fieldvalue.QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(qr.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(qr.Records))
	}
}

func TestHandleQuery_ByRecordID(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"
	setupSchemaForFieldValues(t, s, projectSlug)

	ctx := context.Background()
	rec, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "A",
	}, false, 1)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /field-values/query", HandleQueryFieldValues(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	filter := fieldvalue.FilterNode{Op: "eq", Field: strPtr("record_id"), Value: rec.RecordId.String()}
	body, _ := json.Marshal(fieldvalue.QueryRequest{SchemaKey: "task", Filter: &filter})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/field-values/query?project_slug="+projectSlug, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var qr fieldvalue.QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(qr.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(qr.Records))
	}
}

func TestHandleQuery_RangeAndLike(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"
	setupSchemaForFieldValues(t, s, projectSlug)

	ctx := context.Background()
	if _, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "Fix bug", "priority": float64(1),
	}, false, 1); err != nil {
		t.Fatalf("seed 1: %v", err)
	}
	if _, err := s.MutateRecord(ctx, projectSlug, "task", nil, map[string]any{
		"title": "Add feature", "priority": float64(5),
	}, false, 1); err != nil {
		t.Fatalf("seed 2: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /field-values/query", HandleQueryFieldValues(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	filter := fieldvalue.FilterNode{
		Op: "and",
		Conditions: &[]fieldvalue.FilterNode{
			{Op: "gt", Field: strPtr("priority"), Value: float64(2)},
			{Op: "like", Field: strPtr("title"), Value: "feature"},
		},
	}
	body, _ := json.Marshal(fieldvalue.QueryRequest{SchemaKey: "task", Filter: &filter})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/field-values/query?project_slug="+projectSlug, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var qr fieldvalue.QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(qr.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(qr.Records))
	}
	if qr.Records[0].Values["title"] != "Add feature" {
		t.Fatalf("unexpected title: %v", qr.Records[0].Values["title"])
	}
}

func TestHandleQuery_InvalidFilterOp(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"
	setupSchemaForFieldValues(t, s, projectSlug)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /field-values/query", HandleQueryFieldValues(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	filter := fieldvalue.FilterNode{Op: "invalid", Field: strPtr("priority"), Value: float64(1)}
	body, _ := json.Marshal(fieldvalue.QueryRequest{SchemaKey: "task", Filter: &filter})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/field-values/query?project_slug="+projectSlug, bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleDefinition(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"
	setupSchemaForFieldValues(t, s, projectSlug)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("GET /field-values/definition", HandleFieldValueDefinition(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/field-values/definition?project_slug=" + projectSlug + "&schema_key=task")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var def fieldvalue.SchemaDefinition
	if err := json.NewDecoder(resp.Body).Decode(&def); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if def.SchemaKey != "task" {
		t.Fatalf("schema_key = %q, want %q", def.SchemaKey, "task")
	}
	if len(def.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(def.Fields))
	}
}

func TestHandleDefinition_SchemaNotFound(t *testing.T) {
	s := setupFieldValueStore(t)
	projectSlug := "test-proj"

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("GET /field-values/definition", HandleFieldValueDefinition(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/field-values/definition?project_slug=" + projectSlug + "&schema_key=nonexistent")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	var er map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if er["error"] != "schema_not_found" {
		t.Fatalf("error = %q, want %q", er["error"], "schema_not_found")
	}
}

func strPtr(s string) *string {
	return &s
}
