package agent

import "time"

// Default generation config values.
const (
	DefaultTemperature  = 0.95
	DefaultTopP         = 0.95
	DefaultTopK         = 40
	DefaultMaxToolCalls = 10
)

// MaxReasoningEffortLength is the longest reasoning effort value that can be
// stored. It matches the width of the reasoning_effort column.
const MaxReasoningEffortLength = 64

// Agent represents an AI agent configuration.
type Agent struct {
	ID                     string
	Name                   string
	ProviderID             string
	ModelID                string
	SystemPromptID         string // empty string = no prompt
	TemperatureEnabled     bool
	Temperature            float64
	TopPEnabled            bool
	TopP                   float64
	TopKEnabled            bool
	TopK                   int
	ReasoningEffortEnabled bool
	ReasoningEffort        string // provider-specific value, for example "low" or "2048"
	MaxToolCalls           int
	GapReportingEnabled    bool
	CreatedAt              time.Time
	UpdatedAt              time.Time

	// Resolved fields (populated by JOINs, not persisted directly)
	ProviderName     string
	ProviderType     string
	SystemPromptName string
}
