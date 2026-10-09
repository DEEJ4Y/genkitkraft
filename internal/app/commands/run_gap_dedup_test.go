package commands

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

func TestParseGapDedupDecision(t *testing.T) {
	tests := []struct {
		name    string
		resp    string
		want    gapDedupDecision
		wantErr bool
	}{
		{
			name: "valid JSON",
			resp: `{"action":"create","category":"knowledge","context":"c","details":"d","suggested_resolution":"r"}`,
			want: gapDedupDecision{Action: "create", Category: "knowledge", Context: "c", Details: "d", SuggestedResolution: "r"},
		},
		{
			name: "JSON wrapped in a markdown fence and commentary",
			resp: "Here is my decision:\n```json\n{\"action\":\"merge\",\"gap_id\":\"gap-1\",\"category\":\"capability\",\"context\":\"c\",\"details\":\"d\"}\n```\nHope that helps!",
			want: gapDedupDecision{Action: "merge", GapID: "gap-1", Category: "capability", Context: "c", Details: "d"},
		},
		{
			name:    "missing required fields",
			resp:    `{"action":"create","category":"knowledge"}`,
			wantErr: true,
		},
		{
			name:    "no JSON object at all",
			resp:    "I cannot decide right now.",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGapDedupDecision(tt.resp)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseGapDedupDecision(%q) = %+v, nil, want an error", tt.resp, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseGapDedupDecision(%q): %v", tt.resp, err)
			}
			if got != tt.want {
				t.Errorf("parseGapDedupDecision(%q) = %+v, want %+v", tt.resp, got, tt.want)
			}
		})
	}
}

// dedupTestEnv wires a RunGapDedupCommand against fakes, configurable per test.
type dedupTestEnv struct {
	gapRepo        *mock.GapRepository
	agentRepo      *fakeAgentRepo
	providerRepo   *fakeProviderRepo
	enc            *fakeEncryptor
	playgroundRepo *fakePlaygroundRepo
	chatProvider   *mock.ChatProvider
	cmd            *RunGapDedupCommand
}

func newDedupTestEnv() *dedupTestEnv {
	apiKey := "encrypted-key"
	env := &dedupTestEnv{
		gapRepo: &mock.GapRepository{},
		agentRepo: &fakeAgentRepo{getByIDResult: &agent.Agent{
			ID: "agent-1", ProviderID: "provider-1", ModelID: "gpt-4o-mini",
		}},
		providerRepo: &fakeProviderRepo{getByIDResult: &provider.Provider{
			ID: "provider-1", ProviderType: provider.OpenAI, APIKey: &apiKey, BaseURL: "https://api.openai.com/v1",
		}},
		enc: &fakeEncryptor{decrypted: "sk-test"},
		// Default to a resolvable message: a real PlaygroundRepository never
		// returns (nil, nil) — nil error implies a non-nil message — so tests
		// that don't care about message resolution shouldn't have to set this
		// just to avoid a nil dereference. Tests exercising resolution
		// failure override getLatestMessageErr explicitly.
		playgroundRepo: &fakePlaygroundRepo{getLatestMessageResult: &playground.Message{ID: "message-default"}},
		chatProvider:   &mock.ChatProvider{},
	}
	env.cmd = NewRunGapDedupCommand(env.gapRepo, env.agentRepo, env.providerRepo, env.enc, env.playgroundRepo, env.chatProvider, zerolog.Nop())
	return env
}

func TestRunGapDedupCommand_CreateDecision(t *testing.T) {
	env := newDedupTestEnv()
	env.chatProvider.ChatResponse = `{"action":"create","category":"knowledge","context":"merged context","details":"merged details","suggested_resolution":"add a source"}`
	env.playgroundRepo.getLatestMessageResult = &playground.Message{ID: "message-1"}

	params := RunGapDedupParams{
		SessionID: "session-1", AgentID: "agent-1",
		Category: "knowledge", Context: "raw context", Details: "raw details",
	}
	if err := env.cmd.Execute(context.Background(), params); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if env.gapRepo.LastCreate == nil {
		t.Fatal("gapRepo.Create was not called")
	}
	if env.gapRepo.LastUpdate != nil {
		t.Error("gapRepo.Update was called, want only Create for a 'create' decision")
	}
	if env.gapRepo.LastCreate.Context != "merged context" || env.gapRepo.LastCreate.Details != "merged details" {
		t.Errorf("created gap = %+v, want the model's context/details", env.gapRepo.LastCreate)
	}
	if env.gapRepo.LastAddReference == nil {
		t.Fatal("gapRepo.AddReference was not called")
	}
	if env.gapRepo.LastAddReference.MessageID != "message-1" {
		t.Errorf("reference.MessageID = %q, want the resolved message ID", env.gapRepo.LastAddReference.MessageID)
	}
}

func TestRunGapDedupCommand_MergeDecision_UpdatesExistingOpenGap(t *testing.T) {
	env := newDedupTestEnv()
	existing := &gap.Gap{ID: "gap-existing", AgentID: "agent-1", Status: gap.StatusOpen, Category: gap.CategoryKnowledge}
	env.gapRepo.GetByIDResult = existing
	env.chatProvider.ChatResponse = `{"action":"merge","gap_id":"gap-existing","category":"knowledge","context":"combined context","details":"combined details"}`
	env.playgroundRepo.getLatestMessageResult = &playground.Message{ID: "message-2"}

	params := RunGapDedupParams{SessionID: "session-1", AgentID: "agent-1", Category: "knowledge", Context: "c", Details: "d"}
	if err := env.cmd.Execute(context.Background(), params); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if env.gapRepo.LastCreate != nil {
		t.Error("gapRepo.Create was called, want only Update for a 'merge' decision")
	}
	if env.gapRepo.LastUpdate == nil {
		t.Fatal("gapRepo.Update was not called")
	}
	if env.gapRepo.LastUpdate.ID != "gap-existing" {
		t.Errorf("updated gap ID = %q, want %q", env.gapRepo.LastUpdate.ID, "gap-existing")
	}
	if env.gapRepo.LastUpdate.Status != gap.StatusOpen {
		t.Errorf("merged gap status = %q, want forced to %q", env.gapRepo.LastUpdate.Status, gap.StatusOpen)
	}
	if env.gapRepo.LastAddReference.GapID != "gap-existing" {
		t.Errorf("reference.GapID = %q, want %q", env.gapRepo.LastAddReference.GapID, "gap-existing")
	}
}

// A new report on a dismissed gap reopens it but keeps the dismissal history.
func TestRunGapDedupCommand_MergeDecision_DismissedGap_KeepsDismissalHistory(t *testing.T) {
	env := newDedupTestEnv()
	env.gapRepo.GetByIDResult = &gap.Gap{
		ID: "gap-d", AgentID: "agent-1", Status: gap.StatusDismissed,
		DismissalCategory: gap.DismissalInsufficientDetail, DismissalReason: "need scope",
	}
	env.chatProvider.ChatResponse = `{"action":"merge","gap_id":"gap-d","category":"knowledge","context":"c","details":"d"}`

	params := RunGapDedupParams{SessionID: "session-1", AgentID: "agent-1", Category: "knowledge", Context: "c", Details: "d"}
	if err := env.cmd.Execute(context.Background(), params); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	u := env.gapRepo.LastUpdate
	if u == nil {
		t.Fatal("gapRepo.Update was not called")
	}
	if u.Status != gap.StatusOpen || u.ReopenedFrom != gap.StatusDismissed || u.ReopenedAt == nil {
		t.Errorf("gap = status %q, reopenedFrom %q, reopenedAt %v; want open, dismissed, set", u.Status, u.ReopenedFrom, u.ReopenedAt)
	}
	if u.DismissalCategory != gap.DismissalInsufficientDetail || u.DismissalReason != "need scope" {
		t.Errorf("dismissal fields = %q / %q, want kept", u.DismissalCategory, u.DismissalReason)
	}
	if u.LastReportedAt.IsZero() {
		t.Error("LastReportedAt was not set")
	}
}

func TestRunGapDedupCommand_MergeDecision_ResolvedGap_RecordsResolved(t *testing.T) {
	env := newDedupTestEnv()
	env.gapRepo.GetByIDResult = &gap.Gap{ID: "gap-r", AgentID: "agent-1", Status: gap.StatusResolved}
	env.chatProvider.ChatResponse = `{"action":"merge","gap_id":"gap-r","category":"knowledge","context":"c","details":"d"}`

	params := RunGapDedupParams{AgentID: "agent-1", Category: "knowledge", Context: "c", Details: "d"}
	if err := env.cmd.Execute(context.Background(), params); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if u := env.gapRepo.LastUpdate; u == nil || u.ReopenedFrom != gap.StatusResolved {
		t.Fatalf("LastUpdate = %+v, want reopenedFrom resolved", u)
	}
}

func TestRunGapDedupCommand_MergeDecision_OpenGap_NoReopenHistory(t *testing.T) {
	env := newDedupTestEnv()
	env.gapRepo.GetByIDResult = &gap.Gap{ID: "gap-o", AgentID: "agent-1", Status: gap.StatusOpen}
	env.chatProvider.ChatResponse = `{"action":"merge","gap_id":"gap-o","category":"knowledge","context":"c","details":"d"}`

	params := RunGapDedupParams{AgentID: "agent-1", Category: "knowledge", Context: "c", Details: "d"}
	if err := env.cmd.Execute(context.Background(), params); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	u := env.gapRepo.LastUpdate
	if u == nil || u.ReopenedFrom != "" || u.ReopenedAt != nil || u.LastReportedAt.IsZero() {
		t.Fatalf("LastUpdate = %+v, want no reopen history and LastReportedAt set", u)
	}
}

func TestAgentLocks_SameAgentSerialized_DifferentAgentsIndependent(t *testing.T) {
	var l agentLocks
	unlock, err := l.lock(context.Background(), "agent-1")
	if err != nil {
		t.Fatalf("lock: %v", err)
	}

	got := make(chan struct{})
	go func() {
		u, err := l.lock(context.Background(), "agent-1")
		if err == nil {
			defer u()
		}
		close(got)
	}()
	select {
	case <-got:
		t.Fatal("second lock for the same agent was granted while the first was held")
	case <-time.After(50 * time.Millisecond):
	}

	other, err := l.lock(context.Background(), "agent-2")
	if err != nil {
		t.Fatalf("lock for another agent: %v", err)
	}
	other()

	unlock()
	select {
	case <-got:
	case <-time.After(time.Second):
		t.Fatal("second lock was not granted after unlock")
	}

	// Wait for the goroutine's deferred unlock before the leak check.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		l.mu.Lock()
		n := len(l.entries)
		l.mu.Unlock()
		if n == 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Error("lock entries left behind, want 0")
}

func TestAgentLocks_WaitEndsWhenContextEnds(t *testing.T) {
	var l agentLocks
	unlock, _ := l.lock(context.Background(), "agent-1")
	defer unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := l.lock(ctx, "agent-1"); err == nil {
		t.Fatal("lock returned nil error, want the context error")
	}
}

// countingChatProvider records how many Chat calls run at the same time.
type countingChatProvider struct {
	*mock.ChatProvider
	mu     sync.Mutex
	active int
	max    int
}

func (c *countingChatProvider) Chat(ctx context.Context, req chatprovider.ChatRequest) (string, error) {
	c.mu.Lock()
	c.active++
	if c.active > c.max {
		c.max = c.active
	}
	c.mu.Unlock()
	time.Sleep(30 * time.Millisecond)
	c.mu.Lock()
	c.active--
	c.mu.Unlock()
	return c.ChatProvider.Chat(ctx, req)
}

// #50: two concurrent duplicate reports for one agent must run their dedup
// passes one after the other, so the second pass reads the first pass's gap.
func TestRunGapDedupCommand_ConcurrentReportsSameAgent_RunOneAtATime(t *testing.T) {
	env := newDedupTestEnv()
	env.chatProvider.ChatResponse = `{"action":"create","category":"knowledge","context":"c","details":"d"}`
	counting := &countingChatProvider{ChatProvider: env.chatProvider}
	cmd := NewRunGapDedupCommand(env.gapRepo, env.agentRepo, env.providerRepo, env.enc, env.playgroundRepo, counting, zerolog.Nop())

	params := RunGapDedupParams{AgentID: "agent-1", Category: "knowledge", Context: "c", Details: "d"}
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cmd.Execute(context.Background(), params)
		}()
	}
	wg.Wait()

	if counting.max != 1 {
		t.Errorf("max concurrent dedup passes = %d, want 1", counting.max)
	}
}

// A merge decision naming a terminal gap (dismissed as unrelated) must not
// reopen it — the dedup pipeline should fall through to creating a new gap
// instead, per gap.Gap.IsTerminal.
func TestRunGapDedupCommand_MergeDecision_TerminalTarget_FallsThroughToCreate(t *testing.T) {
	env := newDedupTestEnv()
	terminal := &gap.Gap{ID: "gap-terminal", AgentID: "agent-1", Status: gap.StatusDismissed, DismissalCategory: gap.DismissalUnrelated}
	env.gapRepo.GetByIDResult = terminal
	env.chatProvider.ChatResponse = `{"action":"merge","gap_id":"gap-terminal","category":"knowledge","context":"c","details":"d"}`

	params := RunGapDedupParams{SessionID: "session-1", AgentID: "agent-1", Category: "knowledge", Context: "c", Details: "d"}
	if err := env.cmd.Execute(context.Background(), params); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if env.gapRepo.LastUpdate != nil {
		t.Error("gapRepo.Update was called on a terminal gap, want it left untouched")
	}
	if env.gapRepo.LastCreate == nil {
		t.Fatal("gapRepo.Create was not called — merge into a terminal gap should fall through to create")
	}
}

func TestRunGapDedupCommand_MalformedModelOutput_FallsBackToRawReport(t *testing.T) {
	env := newDedupTestEnv()
	env.chatProvider.ChatResponse = "I'm not sure, sorry."

	params := RunGapDedupParams{
		SessionID: "session-1", AgentID: "agent-1",
		Category: "improvement", Context: "raw context", Details: "raw details", SuggestedResolution: "raw suggestion",
	}
	if err := env.cmd.Execute(context.Background(), params); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if env.gapRepo.LastCreate == nil {
		t.Fatal("gapRepo.Create was not called")
	}
	g := env.gapRepo.LastCreate
	if g.Category != gap.CategoryImprovement || g.Context != "raw context" || g.Details != "raw details" || g.SuggestedResolution != "raw suggestion" {
		t.Errorf("created gap = %+v, want the raw report's fields as a fallback", g)
	}
}

// On the non-streaming chat path, the assistant message isn't persisted
// until after generation finishes, so the pipeline can genuinely never
// resolve a message ID in time — this must not be a hard error.
func TestRunGapDedupCommand_MessageResolutionExhausted_ReferenceHasEmptyMessageID(t *testing.T) {
	env := newDedupTestEnv()
	env.chatProvider.ChatResponse = `{"action":"create","category":"knowledge","context":"c","details":"d"}`
	env.playgroundRepo.getLatestMessageErr = apperrors.NewAppError(apperrors.NotFound, "playground message not found")

	params := RunGapDedupParams{SessionID: "session-1", AgentID: "agent-1", Category: "knowledge", Context: "c", Details: "d"}
	if err := env.cmd.Execute(context.Background(), params); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if env.gapRepo.LastAddReference == nil {
		t.Fatal("gapRepo.AddReference was not called")
	}
	if env.gapRepo.LastAddReference.MessageID != "" {
		t.Errorf("reference.MessageID = %q, want empty when resolution is exhausted", env.gapRepo.LastAddReference.MessageID)
	}
}

// A report from the stateless deploy chat-completions endpoint has no
// session, so there's nothing to poll a message ID against — the pipeline
// must skip resolution entirely (not just resolve to empty) and still write
// a reference, with both SessionID and MessageID empty.
func TestRunGapDedupCommand_EmptySessionID_SkipsMessageResolution_ReferenceHasEmptySessionAndMessageID(t *testing.T) {
	env := newDedupTestEnv()
	env.chatProvider.ChatResponse = `{"action":"create","category":"knowledge","context":"c","details":"d"}`

	params := RunGapDedupParams{SessionID: "", AgentID: "agent-1", Category: "knowledge", Context: "c", Details: "d"}
	if err := env.cmd.Execute(context.Background(), params); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if env.playgroundRepo.getLatestMessageCalls != 0 {
		t.Errorf("GetLatestMessageBySession called %d times, want 0 when SessionID is empty", env.playgroundRepo.getLatestMessageCalls)
	}
	if env.gapRepo.LastAddReference == nil {
		t.Fatal("gapRepo.AddReference was not called")
	}
	if env.gapRepo.LastAddReference.SessionID != "" {
		t.Errorf("reference.SessionID = %q, want empty", env.gapRepo.LastAddReference.SessionID)
	}
	if env.gapRepo.LastAddReference.MessageID != "" {
		t.Errorf("reference.MessageID = %q, want empty", env.gapRepo.LastAddReference.MessageID)
	}
}
