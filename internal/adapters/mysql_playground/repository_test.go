//go:build integration

package mysqlplayground_test

import (
	"context"
	"database/sql"
	"testing"

	mysqldb "github.com/DEEJ4Y/genkitkraft/internal/adapters/mysql_db"
	mysqlplayground "github.com/DEEJ4Y/genkitkraft/internal/adapters/mysql_playground"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	"github.com/DEEJ4Y/genkitkraft/resources/test/containers"
	"github.com/google/uuid"
)

func TestPlaygroundRepositoryMySQL(t *testing.T) {
	dsn := containers.StartMySQLDSN(t)
	testPlaygroundRepository(t, mysqldb.Open, dsn)
}

func TestPlaygroundRepositoryMariaDB(t *testing.T) {
	dsn := containers.StartMariaDBDSN(t)
	testPlaygroundRepository(t, mysqldb.Open, dsn)
}

func testPlaygroundRepository(t *testing.T, open func(string) (*sql.DB, error), dsn string) {
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

	repo := mysqlplayground.NewPlaygroundRepository(db)

	// Create session
	s := &playground.Session{AgentID: agentID, Title: "Test Session"}
	if err := repo.CreateSession(ctx, s); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if s.ID == "" {
		t.Fatal("CreateSession did not assign ID")
	}
	t.Cleanup(func() { repo.DeleteSession(ctx, s.ID) })

	t.Run("GetSession", func(t *testing.T) {
		got, err := repo.GetSession(ctx, s.ID)
		if err != nil {
			t.Fatalf("GetSession: %v", err)
		}
		if got.Title != s.Title || got.AgentID != s.AgentID {
			t.Errorf("fields mismatch: got %+v", got)
		}
	})

	t.Run("ListSessionsByAgent", func(t *testing.T) {
		sessions, err := repo.ListSessionsByAgent(ctx, agentID)
		if err != nil {
			t.Fatalf("ListSessionsByAgent: %v", err)
		}
		if len(sessions) == 0 {
			t.Error("expected at least one session")
		}
	})

	t.Run("UpdateSessionTitle", func(t *testing.T) {
		if err := repo.UpdateSessionTitle(ctx, s.ID, "Updated Title"); err != nil {
			t.Fatalf("UpdateSessionTitle: %v", err)
		}
		got, err := repo.GetSession(ctx, s.ID)
		if err != nil {
			t.Fatalf("GetSession after title update: %v", err)
		}
		if got.Title != "Updated Title" {
			t.Errorf("Title: got %q, want %q", got.Title, "Updated Title")
		}
	})

	t.Run("CreateAndListMessages", func(t *testing.T) {
		m := &playground.Message{SessionID: s.ID, Role: "user", Content: "hello"}
		if err := repo.CreateMessage(ctx, m); err != nil {
			t.Fatalf("CreateMessage: %v", err)
		}
		if m.ID == "" {
			t.Fatal("CreateMessage did not assign ID")
		}

		messages, err := repo.ListMessagesBySession(ctx, s.ID)
		if err != nil {
			t.Fatalf("ListMessagesBySession: %v", err)
		}
		if len(messages) == 0 {
			t.Error("expected at least one message")
		}
		if messages[0].Content != "hello" {
			t.Errorf("Content: got %q, want %q", messages[0].Content, "hello")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		if _, err := repo.GetSession(ctx, "nonexistent-id"); err == nil {
			t.Error("expected error for nonexistent session, got nil")
		}
	})

	t.Run("DeleteSession", func(t *testing.T) {
		if err := repo.DeleteSession(ctx, s.ID); err != nil {
			t.Fatalf("DeleteSession: %v", err)
		}
		if _, err := repo.GetSession(ctx, s.ID); err == nil {
			t.Error("expected error after delete, got nil")
		}
	})
}
