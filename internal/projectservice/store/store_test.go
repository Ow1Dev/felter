package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Ow1Dev/felter/internal/testutil"
)

func TestPostgresStore(t *testing.T) {
	pool := dbtest.StartPostgres(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s := NewPostgresStore(pool)

	t.Run("create and get by slug", func(t *testing.T) {
		name := fmt.Sprintf("project-%d", time.Now().UnixNano())

		created, err := s.CreateProject(ctx, name, "desc")
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.Id == (created.Id) && created.Slug == "" {
			t.Fatal("expected populated project")
		}
		if created.Name != name {
			t.Fatalf("name = %q, want %q", created.Name, name)
		}

		found, err := s.GetProjectBySlug(ctx, created.Slug)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if found.Id != created.Id {
			t.Fatal("id mismatch")
		}
	})

	t.Run("create rejects empty slug", func(t *testing.T) {
		_, err := s.CreateProject(ctx, "!!!", "desc")
		if err != ErrInvalidProjectName {
			t.Fatalf("expected ErrInvalidProjectName, got %v", err)
		}
	})

	t.Run("list projects", func(t *testing.T) {
		name := fmt.Sprintf("list-project-%d", time.Now().UnixNano())

		created, err := s.CreateProject(ctx, name, "")
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		list, err := s.ListProjects(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}

		var found bool
		for _, p := range list {
			if p.Id == created.Id {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("created project not in list")
		}
	})

	t.Run("get project not found", func(t *testing.T) {
		_, err := s.GetProjectBySlug(ctx, "nonexistent-slug-12345")
		if err != ErrProjectNotFound {
			t.Fatalf("expected ErrProjectNotFound, got %v", err)
		}
	})

	t.Run("unique slug collision", func(t *testing.T) {
		base := fmt.Sprintf("collision-%d", time.Now().UnixNano())

		p1, err := s.CreateProject(ctx, base, "")
		if err != nil {
			t.Fatalf("create first: %v", err)
		}
		if p1.Slug != base {
			t.Fatalf("first slug = %q, want %q", p1.Slug, base)
		}

		p2, err := s.CreateProject(ctx, base, "")
		if err != nil {
			t.Fatalf("create second: %v", err)
		}
		wantSlug := base + "-1"
		if p2.Slug != wantSlug {
			t.Fatalf("second slug = %q, want %q", p2.Slug, wantSlug)
		}
	})
}
