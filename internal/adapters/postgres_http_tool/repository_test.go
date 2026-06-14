//go:build integration

package postgreshttptool_test

import (
	"context"
	"testing"

	postgresdb "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_db"
	postgreshttptool "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_http_tool"
	httptool "github.com/DEEJ4Y/genkitkraft/internal/domain/http_tool"
	"github.com/DEEJ4Y/genkitkraft/resources/test/containers"
	"github.com/google/uuid"
)

func TestHttpToolRepositoryPostgres(t *testing.T) {
	url := containers.StartPostgresDSN(t)

	db, err := postgresdb.Open(url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := postgreshttptool.NewHttpToolRepository(db)
	ctx := context.Background()

	tool := &httptool.HttpTool{
		Name:        "test-tool-" + uuid.New().String(),
		Description: "test description",
		Method:      "GET",
		URL:         "https://api.test.com/endpoint",
		Headers:     []httptool.HttpToolHeader{},
	}
	if err := repo.Create(ctx, tool); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tool.ID == "" {
		t.Fatal("Create did not assign ID")
	}
	t.Cleanup(func() { repo.Delete(ctx, tool.ID) })

	t.Run("GetByID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, tool.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != tool.Name || got.Method != tool.Method {
			t.Errorf("fields mismatch: got %+v", got)
		}
	})

	t.Run("Count", func(t *testing.T) {
		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if count == 0 {
			t.Error("expected count > 0")
		}
	})

	t.Run("List", func(t *testing.T) {
		list, err := repo.List(ctx, 10, 0)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(list) == 0 {
			t.Error("expected at least one result")
		}
	})

	t.Run("Update", func(t *testing.T) {
		tool.Description = "updated description"
		if err := repo.Update(ctx, tool); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, err := repo.GetByID(ctx, tool.ID)
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if got.Description != tool.Description {
			t.Errorf("Description after update: got %q, want %q", got.Description, tool.Description)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		if _, err := repo.GetByID(ctx, "nonexistent-id"); err == nil {
			t.Error("expected error for nonexistent ID, got nil")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := repo.Delete(ctx, tool.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.GetByID(ctx, tool.ID); err == nil {
			t.Error("expected error after delete, got nil")
		}
	})
}
