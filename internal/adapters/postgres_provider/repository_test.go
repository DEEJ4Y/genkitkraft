//go:build integration

package postgresperson_test

import (
	"context"
	"os"
	"testing"

	postgresdb "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_db"
	postgresperson "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_provider"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
)

func TestProviderRepositoryPostgres(t *testing.T) {
	url := os.Getenv("TEST_POSTGRES_URL")
	if url == "" {
		t.Skip("TEST_POSTGRES_URL not set")
	}

	db, err := postgresdb.Open(url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := postgresperson.NewProviderRepository(db)
	ctx := context.Background()

	// Clean up any stale test providers from previous interrupted runs.
	db.ExecContext(ctx, "DELETE FROM providers WHERE provider_type = $1", string(provider.OpenAI))

	pt := provider.OpenAI
	p := &provider.Provider{
		Name:         "Test Provider",
		ProviderType: pt,
		BaseURL:      "https://api.test.com",
		Enabled:      true,
	}
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
		if got.Name != p.Name || got.BaseURL != p.BaseURL {
			t.Errorf("fields mismatch: got %+v", got)
		}
		if !got.Enabled {
			t.Error("expected Enabled=true")
		}
	})

	t.Run("GetByType", func(t *testing.T) {
		got, err := repo.GetByType(ctx, pt)
		if err != nil {
			t.Fatalf("GetByType: %v", err)
		}
		if got.ID != p.ID {
			t.Errorf("ID mismatch: got %q, want %q", got.ID, p.ID)
		}
	})

	t.Run("List", func(t *testing.T) {
		list, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(list) == 0 {
			t.Error("expected at least one result")
		}
	})

	t.Run("UniqueConstraint", func(t *testing.T) {
		dup := &provider.Provider{Name: "dup", ProviderType: pt, BaseURL: ""}
		if err := repo.Create(ctx, dup); err == nil {
			repo.Delete(ctx, dup.ID)
			t.Error("expected conflict error for duplicate provider type, got nil")
		}
	})

	t.Run("Update", func(t *testing.T) {
		p.BaseURL = "https://api.updated.com"
		p.Enabled = false
		if err := repo.Update(ctx, p); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, err := repo.GetByID(ctx, p.ID)
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if got.BaseURL != p.BaseURL {
			t.Errorf("BaseURL after update: got %q, want %q", got.BaseURL, p.BaseURL)
		}
		if got.Enabled {
			t.Error("expected Enabled=false after update")
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
