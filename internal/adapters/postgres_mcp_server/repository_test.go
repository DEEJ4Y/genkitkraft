//go:build integration

package postgresmcpserver_test

import (
	"context"
	"os"
	"testing"

	postgresdb "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_db"
	postgresmcpserver "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_mcp_server"
	mcpserver "github.com/DEEJ4Y/genkitkraft/internal/domain/mcp_server"
	"github.com/google/uuid"
)

func TestMcpServerRepositoryPostgres(t *testing.T) {
	url := os.Getenv("TEST_POSTGRES_URL")
	if url == "" {
		t.Skip("TEST_POSTGRES_URL not set")
	}

	db, err := postgresdb.Open(url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := postgresmcpserver.NewMcpServerRepository(db)
	ctx := context.Background()

	srv := &mcpserver.McpServer{
		Name:      "test-mcp-" + uuid.New().String(),
		Transport: "sse",
		URL:       "https://mcp.test.com",
		Headers:   []mcpserver.McpServerHeader{},
	}
	if err := repo.Create(ctx, srv); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if srv.ID == "" {
		t.Fatal("Create did not assign ID")
	}
	t.Cleanup(func() { repo.Delete(ctx, srv.ID) })

	t.Run("GetByID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, srv.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != srv.Name || got.URL != srv.URL {
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
		srv.URL = "https://mcp.updated.com"
		if err := repo.Update(ctx, srv); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, err := repo.GetByID(ctx, srv.ID)
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if got.URL != srv.URL {
			t.Errorf("URL after update: got %q, want %q", got.URL, srv.URL)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		if _, err := repo.GetByID(ctx, "nonexistent-id"); err == nil {
			t.Error("expected error for nonexistent ID, got nil")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := repo.Delete(ctx, srv.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.GetByID(ctx, srv.ID); err == nil {
			t.Error("expected error after delete, got nil")
		}
	})
}
