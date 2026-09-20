//go:build integration

package postgresgap_test

import (
	"context"
	"database/sql"
	"testing"

	postgresdb "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_db"
	postgresgap "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_gap"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	"github.com/DEEJ4Y/genkitkraft/resources/test/containers"
	"github.com/google/uuid"
)

func TestGapRepositoryPostgres(t *testing.T) {
	url := containers.StartPostgresDSN(t)

	db, err := postgresdb.Open(url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()

	providerID := seedProvider(t, db, ctx)
	agentID := seedAgent(t, db, ctx, providerID)
	otherAgentID := seedAgent(t, db, ctx, providerID)
	sessionID := seedSession(t, db, ctx, agentID)

	repo := postgresgap.NewGapRepository(db)

	g := &gap.Gap{
		AgentID:  agentID,
		Category: gap.CategoryKnowledge,
		Context:  "what is the refund policy?",
		Details:  "no refund policy document is configured",
	}
	if err := repo.Create(ctx, g); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if g.ID == "" {
		t.Fatal("Create did not assign ID")
	}

	t.Run("GetByID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, g.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Context != g.Context || got.Details != g.Details || got.Category != g.Category {
			t.Errorf("fields mismatch: got %+v", got)
		}
		if got.Status != gap.StatusOpen {
			t.Errorf("default status = %q, want %q", got.Status, gap.StatusOpen)
		}
	})

	t.Run("List_ScopedToAgent", func(t *testing.T) {
		otherGap := &gap.Gap{AgentID: otherAgentID, Category: gap.CategoryCapability, Context: "c", Details: "d"}
		if err := repo.Create(ctx, otherGap); err != nil {
			t.Fatalf("Create other agent's gap: %v", err)
		}

		list, err := repo.List(ctx, agentID, 10, 0)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("List(%q) returned %d gaps, want 1", agentID, len(list))
		}
		if list[0].AgentID != agentID {
			t.Errorf("List(%q) returned a gap from agent %q", agentID, list[0].AgentID)
		}
	})

	t.Run("Count", func(t *testing.T) {
		count, err := repo.Count(ctx, agentID)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if count != 1 {
			t.Errorf("Count(%q) = %d, want 1", agentID, count)
		}
	})

	t.Run("Update", func(t *testing.T) {
		g.Status = gap.StatusDismissed
		g.DismissalCategory = gap.DismissalDuplicate
		g.DismissalReason = "matches an earlier report"
		g.SuggestedResolution = "add a source"
		if err := repo.Update(ctx, g); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, err := repo.GetByID(ctx, g.ID)
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if got.Status != gap.StatusDismissed || got.DismissalCategory != gap.DismissalDuplicate ||
			got.DismissalReason != "matches an earlier report" || got.SuggestedResolution != "add a source" {
			t.Errorf("fields after update: got %+v", got)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		if _, err := repo.GetByID(ctx, "nonexistent-id"); err == nil {
			t.Error("GetByID(nonexistent): want an error, got nil")
		}
		if err := repo.Update(ctx, &gap.Gap{ID: "nonexistent-id"}); err == nil {
			t.Error("Update(nonexistent): want an error, got nil")
		}
	})

	t.Run("AddReferenceAndListReferences", func(t *testing.T) {
		ref := &gap.Reference{GapID: g.ID, SessionID: sessionID}
		if err := repo.AddReference(ctx, ref); err != nil {
			t.Fatalf("AddReference: %v", err)
		}
		if ref.ID == "" {
			t.Fatal("AddReference did not assign ID")
		}

		refs, err := repo.ListReferences(ctx, g.ID)
		if err != nil {
			t.Fatalf("ListReferences: %v", err)
		}
		if len(refs) != 1 {
			t.Fatalf("len(refs) = %d, want 1", len(refs))
		}
		if refs[0].SessionID != sessionID {
			t.Errorf("SessionID = %q, want %q", refs[0].SessionID, sessionID)
		}
		if refs[0].MessageID != "" {
			t.Errorf("MessageID = %q, want empty when no message was given", refs[0].MessageID)
		}
	})

	// A report from the stateless deploy chat-completions endpoint has no
	// session at all — the reference must still round-trip with both
	// SessionID and MessageID empty (NULL in the DB), not error.
	t.Run("AddReferenceWithoutSession", func(t *testing.T) {
		ref := &gap.Reference{GapID: g.ID}
		if err := repo.AddReference(ctx, ref); err != nil {
			t.Fatalf("AddReference: %v", err)
		}

		refs, err := repo.ListReferences(ctx, g.ID)
		if err != nil {
			t.Fatalf("ListReferences: %v", err)
		}
		var found *gap.Reference
		for _, r := range refs {
			if r.ID == ref.ID {
				found = r
			}
		}
		if found == nil {
			t.Fatalf("ListReferences did not return the session-less reference %q", ref.ID)
		}
		if found.SessionID != "" {
			t.Errorf("SessionID = %q, want empty", found.SessionID)
		}
		if found.MessageID != "" {
			t.Errorf("MessageID = %q, want empty", found.MessageID)
		}
	})
}

// seedProvider inserts a real providers row so agents can satisfy the
// agents.provider_id foreign key.
func seedProvider(t *testing.T, db *sql.DB, ctx context.Context) string {
	t.Helper()
	providerID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO providers (id, name, provider_type, api_key, base_url, config, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, NULL, '', '{}', true, NOW(), NOW())`,
		providerID, "test-provider", "openai")
	if err != nil {
		t.Fatalf("inserting test provider: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM providers WHERE id = $1", providerID) })
	return providerID
}

func seedAgent(t *testing.T, db *sql.DB, ctx context.Context, providerID string) string {
	t.Helper()
	agentID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO agents (id, name, provider_id, model_id, system_prompt_id, temperature_enabled, temperature, top_p_enabled, top_p, top_k_enabled, top_k, max_tool_calls, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NULL, false, 0.7, false, 0.9, false, 40, 10, NOW(), NOW())`,
		agentID, "test-agent-"+agentID, providerID, "gpt-4o")
	if err != nil {
		t.Fatalf("inserting test agent: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM agents WHERE id = $1", agentID) })
	return agentID
}

func seedSession(t *testing.T, db *sql.DB, ctx context.Context, agentID string) string {
	t.Helper()
	sessionID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO playground_sessions (id, agent_id, title, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())`,
		sessionID, agentID, "test-session")
	if err != nil {
		t.Fatalf("inserting test session: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM playground_sessions WHERE id = $1", sessionID) })
	return sessionID
}
