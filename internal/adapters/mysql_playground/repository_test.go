//go:build integration

package mysqlplayground_test

import (
	"context"
	"database/sql"
	"sort"
	"testing"
	"time"

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

	t.Run("ListSessionsByAgentPagedAndCount", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			extra := &playground.Session{AgentID: agentID, Title: "Extra Session"}
			if err := repo.CreateSession(ctx, extra); err != nil {
				t.Fatalf("CreateSession (extra): %v", err)
			}
			t.Cleanup(func() { repo.DeleteSession(ctx, extra.ID) })
		}

		all, err := repo.ListSessionsByAgent(ctx, agentID)
		if err != nil {
			t.Fatalf("ListSessionsByAgent: %v", err)
		}
		total, err := repo.CountSessionsByAgent(ctx, agentID)
		if err != nil {
			t.Fatalf("CountSessionsByAgent: %v", err)
		}
		if total != len(all) || total < 3 {
			t.Fatalf("expected count to match list (%d) and be >= 3, got %d", len(all), total)
		}

		seen := map[string]bool{}
		for offset := 0; offset < total; offset += 2 {
			page, err := repo.ListSessionsByAgentPaged(ctx, agentID, 2, offset)
			if err != nil {
				t.Fatalf("ListSessionsByAgentPaged(offset=%d): %v", offset, err)
			}
			if len(page) == 0 || len(page) > 2 {
				t.Fatalf("offset %d: expected 1-2 sessions, got %d", offset, len(page))
			}
			for _, p := range page {
				if seen[p.ID] {
					t.Errorf("session %s returned on more than one page", p.ID)
				}
				seen[p.ID] = true
			}
		}
		if len(seen) != total {
			t.Errorf("expected pages to cover all %d sessions, covered %d", total, len(seen))
		}

		past, err := repo.ListSessionsByAgentPaged(ctx, agentID, 2, total+10)
		if err != nil {
			t.Fatalf("ListSessionsByAgentPaged (past end): %v", err)
		}
		if len(past) != 0 {
			t.Errorf("expected no sessions past the end, got %d", len(past))
		}

		none, err := repo.CountSessionsByAgent(ctx, uuid.New().String())
		if err != nil || none != 0 {
			t.Errorf("expected 0 sessions for an unknown agent, got %d (err %v)", none, err)
		}
	})

	t.Run("ListSessionsByAgentPagedStableOrderOnEqualUpdatedAt", func(t *testing.T) {
		// Force every session of the agent to the same updated_at so the order is decided only by
		// the tiebreaker (id DESC); pages must then be stable with no gaps or repeats.
		for i := 0; i < 3; i++ {
			tie := &playground.Session{AgentID: agentID, Title: "Tie Session"}
			if err := repo.CreateSession(ctx, tie); err != nil {
				t.Fatalf("CreateSession (tie): %v", err)
			}
			t.Cleanup(func() { repo.DeleteSession(ctx, tie.ID) })
		}

		if _, err := db.ExecContext(ctx,
			`UPDATE playground_sessions SET updated_at = ? WHERE agent_id = ?`,
			time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), agentID); err != nil {
			t.Fatalf("forcing equal updated_at: %v", err)
		}

		full, err := repo.ListSessionsByAgentPaged(ctx, agentID, 100, 0)
		if err != nil {
			t.Fatalf("ListSessionsByAgentPaged (full): %v", err)
		}
		if len(full) < 3 {
			t.Fatalf("expected at least 3 sessions, got %d", len(full))
		}
		fullIDs := make([]string, len(full))
		for i, f := range full {
			fullIDs[i] = f.ID
		}
		if !sort.SliceIsSorted(fullIDs, func(i, j int) bool { return fullIDs[i] > fullIDs[j] }) {
			t.Errorf("expected ids in descending order on equal updated_at, got %v", fullIDs)
		}

		var paged []string
		for offset := 0; offset < len(full); offset++ {
			page, err := repo.ListSessionsByAgentPaged(ctx, agentID, 1, offset)
			if err != nil {
				t.Fatalf("ListSessionsByAgentPaged(limit=1, offset=%d): %v", offset, err)
			}
			if len(page) != 1 {
				t.Fatalf("offset %d: expected exactly 1 session, got %d", offset, len(page))
			}
			paged = append(paged, page[0].ID)
		}
		for i := range fullIDs {
			if paged[i] != fullIDs[i] {
				t.Fatalf("one-at-a-time paging differs from the full page at %d:\npaged: %v\nfull:  %v", i, paged, fullIDs)
			}
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

	t.Run("StreamingMessageLifecycle", func(t *testing.T) {
		msg, err := repo.CreateStreamingMessage(ctx, s.ID)
		if err != nil {
			t.Fatalf("CreateStreamingMessage: %v", err)
		}
		if msg.Status != playground.MessageStatusStreaming {
			t.Errorf("Status: got %q, want %q", msg.Status, playground.MessageStatusStreaming)
		}
		if msg.Content != "" {
			t.Errorf("expected empty initial content, got %q", msg.Content)
		}

		seq1, err := repo.AppendMessageChunk(ctx, msg.ID, "Hello")
		if err != nil {
			t.Fatalf("AppendMessageChunk 1: %v", err)
		}
		if seq1 != 1 {
			t.Errorf("seq1: got %d, want 1", seq1)
		}

		seq2, err := repo.AppendMessageChunk(ctx, msg.ID, " world")
		if err != nil {
			t.Fatalf("AppendMessageChunk 2: %v", err)
		}
		if seq2 != 2 {
			t.Errorf("seq2: got %d, want 2", seq2)
		}

		chunks, err := repo.GetMessageChunksSince(ctx, msg.ID, 0)
		if err != nil {
			t.Fatalf("GetMessageChunksSince: %v", err)
		}
		if len(chunks) != 2 || chunks[0].Content != "Hello" || chunks[1].Content != " world" {
			t.Fatalf("unexpected chunks: %+v", chunks)
		}

		sinceChunks, err := repo.GetMessageChunksSince(ctx, msg.ID, 1)
		if err != nil {
			t.Fatalf("GetMessageChunksSince(since=1): %v", err)
		}
		if len(sinceChunks) != 1 || sinceChunks[0].Seq != 2 {
			t.Fatalf("expected only seq 2 after sinceSeq=1, got %+v", sinceChunks)
		}

		got, err := repo.GetMessage(ctx, msg.ID)
		if err != nil {
			t.Fatalf("GetMessage: %v", err)
		}
		if got.Content != "Hello world" {
			t.Errorf("Content: got %q, want %q", got.Content, "Hello world")
		}

		latest, err := repo.GetLatestMessageBySession(ctx, s.ID)
		if err != nil {
			t.Fatalf("GetLatestMessageBySession: %v", err)
		}
		if latest.ID != msg.ID {
			t.Errorf("GetLatestMessageBySession: got %q, want %q", latest.ID, msg.ID)
		}

		if err := repo.CompleteMessage(ctx, msg.ID); err != nil {
			t.Fatalf("CompleteMessage: %v", err)
		}
		got, err = repo.GetMessage(ctx, msg.ID)
		if err != nil {
			t.Fatalf("GetMessage after complete: %v", err)
		}
		if got.Status != playground.MessageStatusComplete {
			t.Errorf("Status after complete: got %q, want %q", got.Status, playground.MessageStatusComplete)
		}

		// CompleteMessage/FailMessage are guarded by status='streaming', so a
		// second call after the message already reached a terminal status must
		// be a no-op — this is what keeps a slow, late FailMessage from
		// clobbering a message that had already completed successfully.
		if err := repo.FailMessage(ctx, msg.ID); err != nil {
			t.Fatalf("FailMessage (post-complete no-op): %v", err)
		}
		got, err = repo.GetMessage(ctx, msg.ID)
		if err != nil {
			t.Fatalf("GetMessage after no-op FailMessage: %v", err)
		}
		if got.Status != playground.MessageStatusComplete {
			t.Errorf("Status should remain %q after no-op FailMessage, got %q", playground.MessageStatusComplete, got.Status)
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
