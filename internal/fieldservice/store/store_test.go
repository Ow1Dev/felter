package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Ow1Dev/felter/internal/fieldservice/api"
	"github.com/Ow1Dev/felter/internal/testutil"
)

func TestPostgresStore(t *testing.T) {
	pool := dbtest.StartPostgres(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s := NewPostgresStore(pool)

	t.Run("create and get schema with fields", func(t *testing.T) {
		projectSlug := fmt.Sprintf("project-%d", time.Now().UnixNano())

		created, err := s.CreateSchema(ctx, projectSlug, "tasks", "Tasks", 1)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.Key != "tasks" || created.Name != "Tasks" {
			t.Fatalf("unexpected schema: %+v", created)
		}

		_, err = s.CreateSchemaField(ctx, projectSlug, "tasks", "title", api.String, 1)
		if err != nil {
			t.Fatalf("create field: %v", err)
		}

		schema, fields, err := s.GetSchemaWithFields(ctx, projectSlug, "tasks")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if schema.Key != "tasks" {
			t.Fatalf("key mismatch")
		}
		if len(fields) != 1 || fields[0].Key != "title" {
			t.Fatalf("expected 1 field with key title, got %+v", fields)
		}
	})

	t.Run("list schemas by project", func(t *testing.T) {
		projectSlug := fmt.Sprintf("list-project-%d", time.Now().UnixNano())

		_, err := s.CreateSchema(ctx, projectSlug, "orders", "Orders", 1)
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		list, err := s.ListSchemasByProject(ctx, projectSlug)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("expected 1 schema, got %d", len(list))
		}
	})

	t.Run("schema not found", func(t *testing.T) {
		_, _, err := s.GetSchemaWithFields(ctx, "nonexistent", "nonexistent")
		if err != ErrSchemaNotFound {
			t.Fatalf("expected ErrSchemaNotFound, got %v", err)
		}
	})

	t.Run("schema already exists", func(t *testing.T) {
		projectSlug := fmt.Sprintf("dup-project-%d", time.Now().UnixNano())

		_, err := s.CreateSchema(ctx, projectSlug, "dup", "Dup", 1)
		if err != nil {
			t.Fatalf("create first: %v", err)
		}

		_, err = s.CreateSchema(ctx, projectSlug, "dup", "Dup2", 1)
		if err != ErrSchemaAlreadyExists {
			t.Fatalf("expected ErrSchemaAlreadyExists, got %v", err)
		}
	})

	t.Run("delete schema", func(t *testing.T) {
		projectSlug := fmt.Sprintf("del-project-%d", time.Now().UnixNano())

		_, err := s.CreateSchema(ctx, projectSlug, "del", "Del", 1)
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		err = s.DeleteSchema(ctx, projectSlug, "del")
		if err != nil {
			t.Fatalf("delete: %v", err)
		}

		_, _, err = s.GetSchemaWithFields(ctx, projectSlug, "del")
		if err != ErrSchemaNotFound {
			t.Fatalf("expected not found after delete, got %v", err)
		}
	})

	t.Run("field already exists", func(t *testing.T) {
		projectSlug := fmt.Sprintf("fdup-project-%d", time.Now().UnixNano())

		_, err := s.CreateSchema(ctx, projectSlug, "fdup", "FDup", 1)
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		_, err = s.CreateSchemaField(ctx, projectSlug, "fdup", "title", api.String, 1)
		if err != nil {
			t.Fatalf("create field first: %v", err)
		}

		_, err = s.CreateSchemaField(ctx, projectSlug, "fdup", "title", api.Int, 1)
		if err != ErrFieldAlreadyExists {
			t.Fatalf("expected ErrFieldAlreadyExists, got %v", err)
		}
	})

	t.Run("delete field", func(t *testing.T) {
		projectSlug := fmt.Sprintf("fdel-project-%d", time.Now().UnixNano())

		_, err := s.CreateSchema(ctx, projectSlug, "fdel", "FDel", 1)
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		_, err = s.CreateSchemaField(ctx, projectSlug, "fdel", "title", api.String, 1)
		if err != nil {
			t.Fatalf("create field: %v", err)
		}

		err = s.DeleteSchemaField(ctx, projectSlug, "fdel", "title")
		if err != nil {
			t.Fatalf("delete field: %v", err)
		}

		_, fields, err := s.GetSchemaWithFields(ctx, projectSlug, "fdel")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if len(fields) != 0 {
			t.Fatalf("expected 0 fields, got %d", len(fields))
		}
	})
}
