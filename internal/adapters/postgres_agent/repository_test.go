//go:build integration

package postgresagent_test

import (
	"context"
	"testing"

	postgresagent "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_agent"
	postgresdb "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_db"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/resources/test/containers"
	"github.com/google/uuid"
)

func TestAgentRepositoryPostgres(t *testing.T) {
	url := containers.StartPostgresDSN(t)

	db, err := postgresdb.Open(url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()

	// Insert a provider row to satisfy the FK constraint on agents.provider_id.
	providerID := uuid.New().String()
	_, err = db.ExecContext(ctx,
		`INSERT INTO providers (id, name, provider_type, api_key, base_url, config, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, NULL, '', '{}', true, NOW(), NOW())`,
		providerID, "test-provider", "openai")
	if err != nil {
		t.Fatalf("inserting test provider: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM providers WHERE id = $1", providerID) })

	repo := postgresagent.NewAgentRepository(db)

	a := &agent.Agent{
		Name:               "test-agent-" + uuid.New().String(),
		ProviderID:         providerID,
		ModelID:            "claude-sonnet-4-5",
		TemperatureEnabled: true,
		Temperature:        0.7,
		TopPEnabled:        false,
		TopP:               0.9,
		TopKEnabled:        false,
		TopK:               40,
		MaxToolCalls:       10,
	}
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if a.ID == "" {
		t.Fatal("Create did not assign ID")
	}
	t.Cleanup(func() { repo.Delete(ctx, a.ID) })

	t.Run("GetByID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, a.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != a.Name || got.ModelID != a.ModelID {
			t.Errorf("fields mismatch: got %+v", got)
		}
		if !got.TemperatureEnabled {
			t.Error("expected TemperatureEnabled=true")
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
		a.Name = "updated-" + uuid.New().String()
		a.TemperatureEnabled = false
		if err := repo.Update(ctx, a); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, err := repo.GetByID(ctx, a.ID)
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if got.Name != a.Name {
			t.Errorf("Name after update: got %q, want %q", got.Name, a.Name)
		}
		if got.TemperatureEnabled {
			t.Error("expected TemperatureEnabled=false after update")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		if _, err := repo.GetByID(ctx, "nonexistent-id"); err == nil {
			t.Error("expected error for nonexistent ID, got nil")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := repo.Delete(ctx, a.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.GetByID(ctx, a.ID); err == nil {
			t.Error("expected error after delete, got nil")
		}
	})
}
