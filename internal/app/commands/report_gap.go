package commands

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	gapreporter "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_reporter"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

// GapDedupTimeout bounds the detached dedup pipeline goroutine so it can
// never run forever if the LLM call hangs.
const GapDedupTimeout = 2 * time.Minute

// ReportGapCommand implements gapreporter.Reporter. It validates the report
// and launches the dedup pipeline on a context detached from the caller's —
// the live tool call must return immediately regardless of how long dedup
// takes. SessionID is optional: reports from the stateless deploy
// chat-completions endpoint have no session to attach.
type ReportGapCommand struct {
	playgroundRepo playgroundrepo.PlaygroundRepository
	dedup          executors.Executor[RunGapDedupParams]
	logger         zerolog.Logger
}

var _ gapreporter.Reporter = (*ReportGapCommand)(nil)

func NewReportGapCommand(playgroundRepo playgroundrepo.PlaygroundRepository, dedup executors.Executor[RunGapDedupParams], logger zerolog.Logger) *ReportGapCommand {
	return &ReportGapCommand{playgroundRepo: playgroundRepo, dedup: dedup, logger: logger}
}

func (c *ReportGapCommand) Report(ctx context.Context, p gapreporter.ReportParams) error {
	if p.AgentID == "" || p.Category == "" || p.Context == "" || p.Details == "" {
		return errors.NewAppError(errors.InvalidInput, "agent id, category, context, and details are required")
	}

	sessionID := p.SessionID
	if sessionID != "" {
		session, err := c.playgroundRepo.GetSession(ctx, sessionID)
		if err != nil {
			return err
		}
		if session.AgentID != p.AgentID {
			// The session belongs to a different agent than the one reporting —
			// treat this as if no session were given rather than writing an
			// inconsistent reference.
			c.logger.Warn().Str("session_id", sessionID).Str("agent_id", p.AgentID).
				Msg("gap report: session belongs to a different agent, dropping session reference")
			sessionID = ""
		}
	}

	dedupCtx, cancel := context.WithTimeout(context.Background(), GapDedupTimeout)
	go func() {
		defer cancel()
		if err := c.dedup.Execute(dedupCtx, RunGapDedupParams{
			SessionID:           sessionID,
			AgentID:             p.AgentID,
			Category:            p.Category,
			Context:             p.Context,
			Details:             p.Details,
			SuggestedResolution: p.SuggestedResolution,
		}); err != nil {
			c.logger.Error().Err(err).Str("agent_id", p.AgentID).Msg("gap dedup pipeline failed")
		}
	}()

	return nil
}
