//go:build integration

package postgresagenttool_test

import (
	"context"
	"testing"

	postgresagenttool "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_agent_tool"
	postgresdb "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_db"
	agenttoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_tool_repo"
	"github.com/DEEJ4Y/genkitkraft/resources/test/containers"
	"github.com/google/uuid"
)

func TestAgentToolRepositoryPostgres(t *testing.T) {
	url := containers.StartPostgresDSN(t)

	db, err := postgresdb.Open(url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()

	// Insert provider and agent to satisfy FK constraints.
	providerID := uuid.New().String()
	agentID := uuid.New().String()
	_, err = db.ExecContext(ctx,
		`INSERT INTO providers (id, name, provider_type, api_key, base_url, config, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, NULL, '', '{}', true, NOW(), NOW())`,
		providerID, "test-provider", "openai")
	if err != nil {
		t.Fatalf("inserting test provider: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO agents (id, name, provider_id, model_id, system_prompt_id, temperature_enabled, temperature, top_p_enabled, top_p, top_k_enabled, top_k, max_tool_calls, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NULL, false, 0.7, false, 0.9, false, 40, 10, NOW(), NOW())`,
		agentID, "test-agent", providerID, "claude-sonnet-4-5")
	if err != nil {
		t.Fatalf("inserting test agent: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM agents WHERE id = $1", agentID)
		db.ExecContext(ctx, "DELETE FROM providers WHERE id = $1", providerID)
	})

	repo := postgresagenttool.NewRepository(db)

	t.Run("GetByAgentID_empty", func(t *testing.T) {
		cfg, err := repo.GetByAgentID(ctx, agentID)
		if err != nil {
			t.Fatalf("GetByAgentID: %v", err)
		}
		if cfg.AgentID != agentID {
			t.Errorf("AgentID: got %q, want %q", cfg.AgentID, agentID)
		}
		if len(cfg.HttpToolIDs) != 0 || len(cfg.McpServers) != 0 || len(cfg.BuiltInToolIDs) != 0 {
			t.Errorf("expected empty config, got %+v", cfg)
		}
	})

	t.Run("SaveAndGet", func(t *testing.T) {
		cfg := agenttoolrepo.AgentToolConfig{
			AgentID:        agentID,
			HttpToolIDs:    []string{},
			McpServers:     []agenttoolrepo.McpServerToolConfig{},
			BuiltInToolIDs: []string{"web_search"},
		}
		if err := repo.Save(ctx, cfg); err != nil {
			t.Fatalf("Save: %v", err)
		}

		got, err := repo.GetByAgentID(ctx, agentID)
		if err != nil {
			t.Fatalf("GetByAgentID after save: %v", err)
		}
		if len(got.BuiltInToolIDs) != 1 || got.BuiltInToolIDs[0] != "web_search" {
			t.Errorf("BuiltInToolIDs: got %v, want [web_search]", got.BuiltInToolIDs)
		}
	})

	t.Run("SaveEmpty", func(t *testing.T) {
		cfg := agenttoolrepo.AgentToolConfig{
			AgentID:        agentID,
			HttpToolIDs:    []string{},
			McpServers:     []agenttoolrepo.McpServerToolConfig{},
			BuiltInToolIDs: []string{},
		}
		if err := repo.Save(ctx, cfg); err != nil {
			t.Fatalf("Save empty: %v", err)
		}
		got, err := repo.GetByAgentID(ctx, agentID)
		if err != nil {
			t.Fatalf("GetByAgentID after empty save: %v", err)
		}
		if len(got.BuiltInToolIDs) != 0 {
			t.Errorf("expected empty BuiltInToolIDs, got %v", got.BuiltInToolIDs)
		}
	})
}
