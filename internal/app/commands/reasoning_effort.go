package commands

import (
	"strings"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
)

// normalizeReasoningEffort trims the value and checks it against the rules for
// the reasoning effort option. The value is provider-specific free text, so
// the only rules are: it must not be empty when the option is enabled, and it
// must fit the column.
func normalizeReasoningEffort(enabled bool, value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) > agent.MaxReasoningEffortLength {
		return "", errors.NewAppErrorf(errors.InvalidInput, "reasoning effort must be at most %d characters", agent.MaxReasoningEffortLength)
	}
	if enabled && value == "" {
		return "", errors.NewAppError(errors.InvalidInput, "reasoning effort is required when reasoning effort is enabled")
	}
	return value, nil
}
