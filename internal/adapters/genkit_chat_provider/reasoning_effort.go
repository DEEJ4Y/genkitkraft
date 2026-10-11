package genkitchatprovider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/firebase/genkit/go/plugins/ollama"
	"google.golang.org/genai"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
)

const (
	// anthropicDefaultMaxTokens is the max_tokens value sent to Anthropic when
	// thinking is off. When thinking is on, the thinking budget is added to it.
	anthropicDefaultMaxTokens int64 = 4096

	// anthropicMinThinkingBudget is the smallest budget_tokens value that the
	// Anthropic API accepts.
	anthropicMinThinkingBudget int64 = 1024
)

// SupportsReasoningEffort reports whether the reasoning effort option reaches
// the provider. Azure AI Foundry and Bedrock plugins ignore a typed config, so
// the option has no effect for them.
func SupportsReasoningEffort(pt provider.ProviderType) bool {
	switch pt {
	case provider.AzureAIFoundry, provider.Bedrock:
		return false
	default:
		return true
	}
}

// googleThinkingConfig maps a reasoning effort value to a Gemini thinking
// config. An integer is a thinking budget in tokens (0 = off, -1 = dynamic).
// Any other text is a thinking level (minimal, low, medium, high).
func googleThinkingConfig(value string) *genai.ThinkingConfig {
	v := strings.TrimSpace(value)
	if n, err := strconv.ParseInt(v, 10, 32); err == nil {
		budget := int32(n)
		return &genai.ThinkingConfig{ThinkingBudget: &budget}
	}
	return &genai.ThinkingConfig{ThinkingLevel: genai.ThinkingLevel(strings.ToUpper(v))}
}

// anthropicThinkingBudget parses the reasoning effort value as a thinking
// budget in tokens.
func anthropicThinkingBudget(value string) (int64, error) {
	v := strings.TrimSpace(value)
	budget, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid reasoning effort %q for Anthropic: use a whole number of thinking tokens, at least %d", v, anthropicMinThinkingBudget)
	}
	if budget < anthropicMinThinkingBudget {
		return 0, fmt.Errorf("invalid reasoning effort %q for Anthropic: thinking tokens must be at least %d", v, anthropicMinThinkingBudget)
	}
	return budget, nil
}

// ollamaThinkOption maps a reasoning effort value to an Ollama think option.
// "true" and "false" switch thinking on or off. Other text is an effort level
// (low, medium, high for GPT-OSS models) that Ollama validates.
func ollamaThinkOption(value string) *ollama.ThinkOption {
	v := strings.ToLower(strings.TrimSpace(value))
	switch v {
	case "true":
		return ollama.ThinkEnabled(true)
	case "false":
		return ollama.ThinkEnabled(false)
	default:
		return ollama.ThinkEffort(v)
	}
}
