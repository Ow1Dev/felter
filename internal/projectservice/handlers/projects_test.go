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

	"github.com/Ow1Dev/felter/internal/projectservice/api"
	"github.com/Ow1Dev/felter/internal/projectservice/store"
	"github.com/Ow1Dev/felter/internal/testutil"
)

func setupTestServer(t *testing.T) (store.Store, *httptest.Server) {
	t.Helper()

	pool := dbtest.StartPostgres(t)
	s := store.NewPostgresStore(pool)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	mux := http.NewServeMux()
	mux.Handle("POST /{$}", HandleCreateProject(logger, s))
	mux.Handle("GET /{$}", HandleListProjects(logger, s))
	mux.Handle("GET /{slug}", HandleGetProject(logger, s))

	return s, httptest.NewServer(mux)
}

func TestHandleCreateProject(t *testing.T) {
	_, srv := setupTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(api.CreateProjectRequest{Name: "test-project", Description: strPtr("desc")})
	resp, err := http.Post(srv.URL+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var project api.Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if project.Name != "test-project" {
		t.Fatalf("name = %q, want %q", project.Name, "test-project")
	}
}

func TestHandleCreateProjectInvalidName(t *testing.T) {
	_, srv := setupTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(api.CreateProjectRequest{Name: "!!!"})
	resp, err := http.Post(srv.URL+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleCreateProjectEmptyName(t *testing.T) {
	_, srv := setupTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(api.CreateProjectRequest{Name: "   "})
	resp, err := http.Post(srv.URL+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleListProjects(t *testing.T) {
	_, srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var projects []api.Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if projects == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestHandleGetProject(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := store.NewPostgresStore(pool)

	ctx := context.Background()
	created, err := s.CreateProject(ctx, "get-test-project", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mux := http.NewServeMux()
	mux.Handle("GET /{slug}", HandleGetProject(logger, s))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/" + created.Slug)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var project api.Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if project.Id != created.Id {
		t.Fatal("id mismatch")
	}
}

func TestHandleGetProjectNotFound(t *testing.T) {
	_, srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/nonexistent-slug-12345")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func strPtr(s string) *string {
	return &s
}
