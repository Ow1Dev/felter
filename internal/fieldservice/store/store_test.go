package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Ow1Dev/felter/internal/testutil"
)

func TestNewPostgresStore(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	if s == nil {
		t.Fatal("expected non-nil store")
	}
	if s.db != pool {
		t.Fatal("expected db to be set")
	}
}

func TestResolveSchemaID_Success(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projectSlug := fmt.Sprintf("resolve-%d", time.Now().UnixNano())
	if _, err := s.CreateSchema(ctx, projectSlug, "test", "Test", 1); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	id, err := s.resolveSchemaID(ctx, projectSlug, "test")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty id")
	}
}

func TestResolveSchemaID_NotFound(t *testing.T) {
	pool := dbtest.StartPostgres(t)
	s := NewPostgresStore(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.resolveSchemaID(ctx, "nonexistent", "nonexistent")
	if err != ErrSchemaNotFound {
		t.Fatalf("expected ErrSchemaNotFound, got %v", err)
	}
}
