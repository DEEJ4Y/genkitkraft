package gap_test

import (
	"testing"

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
