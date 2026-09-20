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

// ReportGapCommand implements gapreporter.Reporter. It validates the report,
// resolves which agent it belongs to, and launches the dedup pipeline on a
// context detached from the caller's — the live tool call must return
// immediately regardless of how long dedup takes.
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
	if p.SessionID == "" || p.Category == "" || p.Context == "" || p.Details == "" {
		return errors.NewAppError(errors.InvalidInput, "session id, category, context, and details are required")
	}

	session, err := c.playgroundRepo.GetSession(ctx, p.SessionID)
	if err != nil {
		return err
	}

	dedupCtx, cancel := context.WithTimeout(context.Background(), GapDedupTimeout)
	go func() {
		defer cancel()
		if err := c.dedup.Execute(dedupCtx, RunGapDedupParams{
			SessionID:           p.SessionID,
			AgentID:             session.AgentID,
			Category:            p.Category,
			Context:             p.Context,
			Details:             p.Details,
			SuggestedResolution: p.SuggestedResolution,
		}); err != nil {
			c.logger.Error().Err(err).Str("session_id", p.SessionID).Msg("gap dedup pipeline failed")
		}
	}()

	return nil
}
