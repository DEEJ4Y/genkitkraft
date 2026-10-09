package genkitchatprovider

import (
	"context"
	"testing"

	gapreporter "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_reporter"
)

// stubReporter captures the last Report() call for assertions, mirroring
// stubCache's shape in builtin_tools_test.go.
type stubReporter struct {
	err error

	called     bool
	lastParams gapreporter.ReportParams
}

func (s *stubReporter) Report(_ context.Context, p gapreporter.ReportParams) error {
	s.called = true
	s.lastParams = p
	return s.err
}

func TestReportGap_ForwardsFieldsToReporter(t *testing.T) {
	reporter := &stubReporter{}
	cp := &ChatProvider{gapReporter: reporter}

	cp.reportGap(context.Background(), "session-1", "agent-1", "capability", "user asked to send an email", "no email tool configured", "add an email tool")

	if !reporter.called {
		t.Fatal("gapReporter.Report was not called")
	}
	want := gapreporter.ReportParams{
		SessionID:           "session-1",
		AgentID:             "agent-1",
		Category:            "capability",
		Context:             "user asked to send an email",
		Details:             "no email tool configured",
		SuggestedResolution: "add an email tool",
	}
	if reporter.lastParams != want {
		t.Errorf("Report called with %+v, want %+v", reporter.lastParams, want)
	}
}

// A broken reporter must never disrupt the live response — reportGap has no
// return value by design, so the only thing to verify is that the call
// still happens and nothing panics.
func TestReportGap_ReporterErrorIsSwallowed(t *testing.T) {
	reporter := &stubReporter{err: context.DeadlineExceeded}
	cp := &ChatProvider{gapReporter: reporter}

	cp.reportGap(context.Background(), "session-1", "agent-1", "knowledge", "c", "d", "")

	if !reporter.called {
		t.Fatal("gapReporter.Report was not called despite the configured error")
	}
}

func TestBuildBuiltInTools_ReportGapGating(t *testing.T) {
	tests := []struct {
		name                string
		gapReportingEnabled bool
		sessionID           string
		agentID             string
		reporter            gapreporter.Reporter
		wantReportGap       bool
	}{
		{"flag disabled", false, "session-1", "agent-1", &stubReporter{}, false},
		{"flag enabled, no session, but agent known (stateless)", true, "", "agent-1", &stubReporter{}, true},
		{"flag enabled but no agentID", true, "session-1", "", &stubReporter{}, false},
		{"flag enabled and agent set but no reporter wired", true, "session-1", "agent-1", nil, false},
		{"flag enabled, session and agent set, reporter wired", true, "session-1", "agent-1", &stubReporter{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := &ChatProvider{gapReporter: tt.reporter}

			tools := cp.buildBuiltInTools(nil, tt.sessionID, tt.agentID, tt.gapReportingEnabled)

			got := false
			names := make([]string, len(tools))
			for i, tool := range tools {
				names[i] = tool.Name()
				if tool.Name() == "report_gap" {
					got = true
				}
			}
			if got != tt.wantReportGap {
				t.Errorf("report_gap present = %v, want %v (tools=%v)", got, tt.wantReportGap, names)
			}
		})
	}
}
