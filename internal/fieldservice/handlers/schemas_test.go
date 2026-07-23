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
	"github.com/Ow1Dev/felter/internal/fieldservice/store"
	"github.com/Ow1Dev/felter/internal/testutil"
)

func setupTestServer(t *testing.T) (*store.PostgresStore, *httptest.Server) {
	t.Helper()

	pool := dbtest.StartPostgres(t)
	s := store.NewPostgresStore(pool)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	mux := http.NewServeMux()
	mux.Handle("GET /schemas", HandleListSchemas(logger, s))
	mux.Handle("POST /schemas", HandleCreateSchema(logger, s))
	mux.Handle("GET /schemas/{schemaKey}", HandleGetSchema(logger, s))
	mux.Handle("DELETE /schemas/{schemaKey}", HandleDeleteSchema(logger, s))
	mux.Handle("POST /schemas/{schemaKey}/fields", HandleCreateSchemaField(logger, s))
	mux.Handle("DELETE /schemas/{schemaKey}/fields/{fieldKey}", HandleDeleteSchemaField(logger, s))

	return s, httptest.NewServer(mux)
}

func TestHandleCreateSchema(t *testing.T) {
	_, srv := setupTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(api.CreateSchemaRequest{ProjectSlug: "test-proj", Key: "tasks", Name: "Tasks"})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/schemas", bytes.NewReader(body))
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

	var schema api.Schema
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if schema.Key != "tasks" {
		t.Fatalf("key = %q, want %q", schema.Key, "tasks")
	}
}

func TestHandleCreateSchemaMissingUserID(t *testing.T) {
	_, srv := setupTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(api.CreateSchemaRequest{ProjectSlug: "test-proj", Key: "tasks", Name: "Tasks"})
	resp, err := http.Post(srv.URL+"/schemas", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestHandleListSchemas(t *testing.T) {
	_, srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/schemas?project_slug=test-proj")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var schemas []api.Schema
	if err := json.NewDecoder(resp.Body).Decode(&schemas); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if schemas == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestHandleGetSchema(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := store.NewPostgresStore(pool)

	ctx := context.Background()
	_, err := s.CreateSchema(ctx, "get-proj", "tasks", "Tasks", 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = s.CreateSchemaField(ctx, "get-proj", "tasks", "title", api.String, 1)
	if err != nil {
		t.Fatalf("create field: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("GET /schemas/{schemaKey}", HandleGetSchema(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/schemas/tasks?project_slug=get-proj")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var schema api.SchemaWithFields
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if schema.Key != "tasks" {
		t.Fatalf("key = %q, want %q", schema.Key, "tasks")
	}
	if schema.Fields == nil || len(*schema.Fields) != 1 {
		t.Fatalf("expected 1 field, got %+v", schema.Fields)
	}
}

func TestHandleGetSchemaNotFound(t *testing.T) {
	_, srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/schemas/nonexistent?project_slug=test-proj")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestHandleDeleteSchema(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := store.NewPostgresStore(pool)

	ctx := context.Background()
	_, err := s.CreateSchema(ctx, "del-proj", "tasks", "Tasks", 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("DELETE /schemas/{schemaKey}", HandleDeleteSchema(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/schemas/tasks?project_slug=del-proj", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestHandleCreateSchemaField(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := store.NewPostgresStore(pool)

	ctx := context.Background()
	_, err := s.CreateSchema(ctx, "field-proj", "tasks", "Tasks", 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("POST /schemas/{schemaKey}/fields", HandleCreateSchemaField(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	body, _ := json.Marshal(api.CreateSchemaFieldRequest{Key: "title", Type: api.String})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/schemas/tasks/fields?project_slug=field-proj", bytes.NewReader(body))
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

	var field api.SchemaField
	if err := json.NewDecoder(resp.Body).Decode(&field); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if field.Key != "title" {
		t.Fatalf("key = %q, want %q", field.Key, "title")
	}
}

func TestHandleDeleteSchemaField(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := store.NewPostgresStore(pool)

	ctx := context.Background()
	_, err := s.CreateSchema(ctx, "fdel-proj", "tasks", "Tasks", 1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = s.CreateSchemaField(ctx, "fdel-proj", "tasks", "title", api.String, 1)
	if err != nil {
		t.Fatalf("create field: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("DELETE /schemas/{schemaKey}/fields/{fieldKey}", HandleDeleteSchemaField(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/schemas/tasks/fields/title?project_slug=fdel-proj", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}
