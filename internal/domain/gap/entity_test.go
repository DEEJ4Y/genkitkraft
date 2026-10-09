package gap_test

import (
	"testing"
	"time"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
)

func TestGapIsTerminal(t *testing.T) {
	tests := []struct {
		name              string
		status            gap.Status
		dismissalCategory string
		want              bool
	}{
		{"dismissed as unrelated is terminal", gap.StatusDismissed, gap.DismissalUnrelated, true},
		{"dismissed as duplicate is not terminal", gap.StatusDismissed, gap.DismissalDuplicate, false},
		{"dismissed as insufficient detail is not terminal", gap.StatusDismissed, gap.DismissalInsufficientDetail, false},
		{"dismissed as other is not terminal", gap.StatusDismissed, gap.DismissalOther, false},
		{"open is never terminal, even with a stray unrelated category", gap.StatusOpen, gap.DismissalUnrelated, false},
		{"resolved is never terminal, even with a stray unrelated category", gap.StatusResolved, gap.DismissalUnrelated, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gap.Gap{Status: tt.status, DismissalCategory: tt.dismissalCategory}
			if got := g.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v (status=%s, dismissalCategory=%s)", got, tt.want, tt.status, tt.dismissalCategory)
			}
		})
	}
}

func TestGapReopen(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)

	t.Run("dismissed gap keeps dismissal history", func(t *testing.T) {
		g := &gap.Gap{Status: gap.StatusDismissed, DismissalCategory: gap.DismissalInsufficientDetail, DismissalReason: "need scope"}
		g.Reopen(now)
		if g.Status != gap.StatusOpen || g.ReopenedFrom != gap.StatusDismissed {
			t.Errorf("status=%q reopenedFrom=%q, want open and dismissed", g.Status, g.ReopenedFrom)
		}
		if g.ReopenedAt == nil || !g.ReopenedAt.Equal(now) {
			t.Errorf("ReopenedAt = %v, want %v", g.ReopenedAt, now)
		}
		if g.DismissalCategory != gap.DismissalInsufficientDetail || g.DismissalReason != "need scope" {
			t.Errorf("dismissal fields changed: %q / %q", g.DismissalCategory, g.DismissalReason)
		}
	})

	t.Run("resolved gap records resolved", func(t *testing.T) {
		g := &gap.Gap{Status: gap.StatusResolved}
		g.Reopen(now)
		if g.Status != gap.StatusOpen || g.ReopenedFrom != gap.StatusResolved {
			t.Errorf("status=%q reopenedFrom=%q, want open and resolved", g.Status, g.ReopenedFrom)
		}
	})

	t.Run("open gap is unchanged", func(t *testing.T) {
		g := &gap.Gap{Status: gap.StatusOpen}
		g.Reopen(now)
		if g.ReopenedFrom != "" || g.ReopenedAt != nil {
			t.Errorf("open gap got reopen history: %q %v", g.ReopenedFrom, g.ReopenedAt)
		}
	})
}
