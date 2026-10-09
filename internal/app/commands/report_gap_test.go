package commands

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	gapreporter "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_reporter"
)

// fakeDedupExecutor implements executors.Executor[RunGapDedupParams]. If
// block is non-nil, Execute waits for it to be closed before recording the
// call — used to prove ReportGapCommand.Report dispatches asynchronously
// rather than waiting on the dedup pipeline.
type fakeDedupExecutor struct {
	block chan struct{}
	err   error

	mu         sync.Mutex
	calls      int
	lastParams RunGapDedupParams
	done       chan struct{}
}

func newFakeDedupExecutor() *fakeDedupExecutor {
	return &fakeDedupExecutor{done: make(chan struct{}, 10)}
}

func (f *fakeDedupExecutor) Execute(_ context.Context, params RunGapDedupParams) error {
	if f.block != nil {
		<-f.block
	}
	f.mu.Lock()
	f.calls++
	f.lastParams = params
	f.mu.Unlock()
	f.done <- struct{}{}
	return f.err
}

func (f *fakeDedupExecutor) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func TestReportGapCommand_MissingFields_ReturnsInvalidInputWithoutDispatching(t *testing.T) {
	dedup := newFakeDedupExecutor()
	cmd := NewReportGapCommand(&fakePlaygroundRepo{}, dedup, zerolog.Nop())

	tests := []gapreporter.ReportParams{
		{Category: "knowledge", Context: "c", Details: "d"},               // missing AgentID
		{AgentID: "a", Context: "c", Details: "d"},                        // missing Category
		{AgentID: "a", Category: "knowledge", Details: "d"},               // missing Context
		{AgentID: "a", Category: "knowledge", Context: "c"},               // missing Details
	}
	for _, p := range tests {
		err := cmd.Report(context.Background(), p)
		appErr, ok := apperrors.IsAppError(err)
		if !ok || appErr.Code() != apperrors.InvalidInput {
			t.Errorf("Report(%+v) = %v, want InvalidInput AppError", p, err)
		}
	}

	if dedup.callCount() != 0 {
		t.Errorf("dedup executor called %d times, want 0 for invalid reports", dedup.callCount())
	}
}

func TestReportGapCommand_UnknownSession_ReturnsErrorWithoutDispatching(t *testing.T) {
	sessionErr := apperrors.NewAppError(apperrors.NotFound, "playground session not found")
	dedup := newFakeDedupExecutor()
	cmd := NewReportGapCommand(&fakePlaygroundRepo{getSessionErr: sessionErr}, dedup, zerolog.Nop())

	err := cmd.Report(context.Background(), gapreporter.ReportParams{
		SessionID: "unknown-session", AgentID: "agent-1", Category: "knowledge", Context: "c", Details: "d",
	})

	appErr, ok := apperrors.IsAppError(err)
	if !ok || appErr.Code() != apperrors.NotFound {
		t.Fatalf("Report() = %v, want the session lookup's NotFound error", err)
	}
	if dedup.callCount() != 0 {
		t.Errorf("dedup executor called %d times, want 0 when the session can't be resolved", dedup.callCount())
	}
}

// The live tool call must return immediately regardless of how long dedup
// takes — this is the whole point of dispatching it on a detached goroutine.
func TestReportGapCommand_ReturnsImmediately_DedupRunsAsync(t *testing.T) {
	dedup := newFakeDedupExecutor()
	dedup.block = make(chan struct{})

	playgroundRepo := &fakePlaygroundRepo{
		getSessionResult: &playground.Session{ID: "session-1", AgentID: "agent-42"},
	}
	cmd := NewReportGapCommand(playgroundRepo, dedup, zerolog.Nop())

	reportErr := cmd.Report(context.Background(), gapreporter.ReportParams{
		SessionID:           "session-1",
		AgentID:             "agent-42",
		Category:            "capability",
		Context:             "user asked to send an email",
		Details:             "no email tool is configured for this agent",
		SuggestedResolution: "add an email-sending HTTP tool",
	})
	if reportErr != nil {
		t.Fatalf("Report() = %v, want nil (dedup dispatch must not block or fail the tool call)", reportErr)
	}
	if dedup.callCount() != 0 {
		t.Fatal("dedup executor ran synchronously — Report() should have returned before it started")
	}

	close(dedup.block)
	select {
	case <-dedup.done:
	case <-time.After(2 * time.Second):
		t.Fatal("dedup executor was never invoked after unblocking")
	}

	got := dedup.lastParams
	want := RunGapDedupParams{
		SessionID:           "session-1",
		AgentID:             "agent-42",
		Category:            "capability",
		Context:             "user asked to send an email",
		Details:             "no email tool is configured for this agent",
		SuggestedResolution: "add an email-sending HTTP tool",
	}
	if got != want {
		t.Errorf("dedup executor called with %+v, want %+v", got, want)
	}
}

// A report from the stateless deploy chat-completions endpoint has no
// session at all — AgentID alone must be enough to dispatch dedup, with no
// session lookup performed.
func TestReportGapCommand_NoSessionID_StillDispatchesUsingAgentID(t *testing.T) {
	dedup := newFakeDedupExecutor()
	dedup.block = make(chan struct{})
	close(dedup.block)

	playgroundRepo := &fakePlaygroundRepo{}
	cmd := NewReportGapCommand(playgroundRepo, dedup, zerolog.Nop())

	reportErr := cmd.Report(context.Background(), gapreporter.ReportParams{
		AgentID:  "agent-42",
		Category: "knowledge",
		Context:  "user asked about a policy",
		Details:  "no source configured for this policy",
	})
	if reportErr != nil {
		t.Fatalf("Report() = %v, want nil for a stateless (no session) report", reportErr)
	}

	select {
	case <-dedup.done:
	case <-time.After(2 * time.Second):
		t.Fatal("dedup executor was never invoked")
	}

	got := dedup.lastParams
	want := RunGapDedupParams{
		SessionID: "",
		AgentID:   "agent-42",
		Category:  "knowledge",
		Context:   "user asked about a policy",
		Details:   "no source configured for this policy",
	}
	if got != want {
		t.Errorf("dedup executor called with %+v, want %+v", got, want)
	}
}

// A session belonging to a different agent than the one reporting must not
// produce an inconsistent reference — the report should still succeed and
// dispatch, but with the session dropped.
func TestReportGapCommand_SessionBelongsToDifferentAgent_DropsSessionReference(t *testing.T) {
	dedup := newFakeDedupExecutor()
	dedup.block = make(chan struct{})
	close(dedup.block)

	playgroundRepo := &fakePlaygroundRepo{
		getSessionResult: &playground.Session{ID: "session-1", AgentID: "some-other-agent"},
	}
	cmd := NewReportGapCommand(playgroundRepo, dedup, zerolog.Nop())

	reportErr := cmd.Report(context.Background(), gapreporter.ReportParams{
		SessionID: "session-1",
		AgentID:   "agent-42",
		Category:  "knowledge",
		Context:   "c",
		Details:   "d",
	})
	if reportErr != nil {
		t.Fatalf("Report() = %v, want nil", reportErr)
	}

	select {
	case <-dedup.done:
	case <-time.After(2 * time.Second):
		t.Fatal("dedup executor was never invoked")
	}

	if dedup.lastParams.SessionID != "" {
		t.Errorf("dedup executor called with SessionID = %q, want empty when the session belongs to a different agent", dedup.lastParams.SessionID)
	}
	if dedup.lastParams.AgentID != "agent-42" {
		t.Errorf("dedup executor called with AgentID = %q, want the reporting agent's ID", dedup.lastParams.AgentID)
	}
}
