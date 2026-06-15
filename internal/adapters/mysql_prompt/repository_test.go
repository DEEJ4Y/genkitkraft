//go:build integration

package mysqlprompt_test

import (
	"context"
	"database/sql"
	"testing"

	mysqldb "github.com/DEEJ4Y/genkitkraft/internal/adapters/mysql_db"
	mysqlprompt "github.com/DEEJ4Y/genkitkraft/internal/adapters/mysql_prompt"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	"github.com/DEEJ4Y/genkitkraft/resources/test/containers"
	"github.com/google/uuid"
)

func TestPromptRepositoryMySQL(t *testing.T) {
	dsn := containers.StartMySQLDSN(t)
	testPromptRepository(t, mysqldb.Open, dsn)
}

func TestPromptRepositoryMariaDB(t *testing.T) {
	dsn := containers.StartMariaDBDSN(t)
	testPromptRepository(t, mysqldb.Open, dsn)
}

func testPromptRepository(t *testing.T, open func(string) (*sql.DB, error), dsn string) {
	t.Helper()

	db, err := open(dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := mysqlprompt.NewPromptRepository(db)
	ctx := context.Background()

	p := &prompt.Prompt{Name: "test-" + uuid.New().String(), Content: "hello world"}
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID == "" {
		t.Fatal("Create did not assign ID")
	}
	t.Cleanup(func() { repo.Delete(ctx, p.ID) })

	t.Run("GetByID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, p.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != p.Name || got.Content != p.Content {
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
		p.Name = "updated-" + uuid.New().String()
		if err := repo.Update(ctx, p); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, err := repo.GetByID(ctx, p.ID)
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if got.Name != p.Name {
			t.Errorf("Name after update: got %q, want %q", got.Name, p.Name)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		if _, err := repo.GetByID(ctx, "nonexistent-id"); err == nil {
			t.Error("expected error for nonexistent ID, got nil")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := repo.Delete(ctx, p.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.GetByID(ctx, p.ID); err == nil {
			t.Error("expected error after delete, got nil")
		}
	})
}
