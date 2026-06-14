//go:build integration

package mysqlagenttool_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	mysqldb "github.com/DEEJ4Y/genkitkraft/internal/adapters/mysql_db"
	mysqlagenttool "github.com/DEEJ4Y/genkitkraft/internal/adapters/mysql_agent_tool"
	agenttoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_tool_repo"
	"github.com/google/uuid"
)

func TestAgentToolRepositoryMySQL(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_URL")
	if dsn == "" {
		t.Skip("TEST_MYSQL_URL not set")
	}
	testAgentToolRepository(t, mysqldb.Open, dsn)
}

func TestAgentToolRepositoryMariaDB(t *testing.T) {
	dsn := os.Getenv("TEST_MARIADB_URL")
	if dsn == "" {
		t.Skip("TEST_MARIADB_URL not set")
	}
	testAgentToolRepository(t, mysqldb.Open, dsn)
}

func testAgentToolRepository(t *testing.T, open func(string) (*sql.DB, error), dsn string) {
	t.Helper()

	db, err := open(dsn)
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
		 VALUES (?, ?, ?, NULL, '', '{}', 1, NOW(), NOW())`,
		providerID, "test-provider", "openai")
	if err != nil {
		t.Fatalf("inserting test provider: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO agents (id, name, provider_id, model_id, system_prompt_id, temperature_enabled, temperature, top_p_enabled, top_p, top_k_enabled, top_k, max_tool_calls, created_at, updated_at)
		 VALUES (?, ?, ?, ?, NULL, 0, 0.7, 0, 0.9, 0, 40, 10, NOW(), NOW())`,
		agentID, "test-agent", providerID, "gpt-4o")
	if err != nil {
		t.Fatalf("inserting test agent: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM agents WHERE id = ?", agentID)
		db.ExecContext(ctx, "DELETE FROM providers WHERE id = ?", providerID)
	})

	repo := mysqlagenttool.NewRepository(db)

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
