package agent

import "time"

// Default generation config values.
const (
	DefaultTemperature = 0.95
	DefaultTopP        = 0.95
	DefaultTopK        = 40
)

// Agent represents an AI agent configuration.
type Agent struct {
	ID                 string
	Name               string
	ProviderID         string
	ModelID            string
	SystemPromptID     string // empty string = no prompt
	TemperatureEnabled bool
	Temperature        float64
	TopPEnabled        bool
	TopP               float64
	TopKEnabled        bool
	TopK               int
	CreatedAt          time.Time
	UpdatedAt          time.Time

	// Resolved fields (populated by JOINs, not persisted directly)
	ProviderName     string
	ProviderType     string
	SystemPromptName string
}
