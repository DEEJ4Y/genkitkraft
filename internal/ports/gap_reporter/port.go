package gapreporter

import "context"

// ReportParams carries a single gap self-report from the live agent.
type ReportParams struct {
	SessionID           string
	Category            string // "knowledge", "capability", or "improvement"
	Context             string
	Details             string
	SuggestedResolution string
}

// Reporter accepts a raw gap report from the report_gap tool and hands it
// off to the background dedup pipeline. Implementations must return quickly
// — the live response must never wait on this call.
type Reporter interface {
	Report(ctx context.Context, p ReportParams) error
}
