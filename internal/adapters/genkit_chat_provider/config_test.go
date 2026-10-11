package genkitchatprovider

import (
	"encoding/json"
	"strings"
	"testing"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/openai/openai-go"
	"google.golang.org/genai"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
)

func reasoningReq(pt provider.ProviderType, enabled bool, value string) chatprovider.ChatRequest {
	return chatprovider.ChatRequest{
		ProviderType:           string(pt),
		ReasoningEffortEnabled: enabled,
		ReasoningEffort:        value,
	}
}

func TestBuildConfig_ReasoningDisabled_KeepsExistingBehavior(t *testing.T) {
	// A stored value must be ignored while the option is disabled.
	for _, pt := range []provider.ProviderType{
		provider.GoogleAI, provider.VertexAI, provider.Ollama,
		provider.OpenAI, provider.XAI, provider.DeepSeek, provider.AzureOpenAI, provider.OpenAICompatible,
	} {
		t.Run(string(pt), func(t *testing.T) {
			cfg, err := buildConfig(reasoningReq(pt, false, "high"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg != nil {
				t.Fatalf("expected an untyped nil config, got %#v", cfg)
			}
		})
	}

	t.Run("anthropic", func(t *testing.T) {
		cfg, err := buildConfig(reasoningReq(provider.Anthropic, false, "2048"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		params, ok := cfg.(*anthropicsdk.MessageNewParams)
		if !ok {
			t.Fatalf("expected *MessageNewParams, got %T", cfg)
		}
		if params.MaxTokens != 4096 {
			t.Errorf("MaxTokens = %d, want 4096", params.MaxTokens)
		}
		if params.Thinking.OfEnabled != nil {
			t.Error("thinking must not be set when reasoning effort is disabled")
		}
	})
}

func TestBuildConfig_Google_Reasoning(t *testing.T) {
	tests := []struct {
		value      string
		wantLevel  genai.ThinkingLevel
		wantBudget *int32
	}{
		{value: "low", wantLevel: genai.ThinkingLevelLow},
		{value: " High ", wantLevel: genai.ThinkingLevelHigh},
		{value: "minimal", wantLevel: genai.ThinkingLevelMinimal},
		{value: "2048", wantBudget: genai.Ptr(int32(2048))},
		{value: "-1", wantBudget: genai.Ptr(int32(-1))},
		{value: "0", wantBudget: genai.Ptr(int32(0))},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			cfg, err := buildConfig(reasoningReq(provider.GoogleAI, true, tt.value))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gc, ok := cfg.(*genai.GenerateContentConfig)
			if !ok {
				t.Fatalf("expected *GenerateContentConfig, got %T", cfg)
			}
			tc := gc.ThinkingConfig
			if tc == nil {
				t.Fatal("ThinkingConfig is nil")
			}
			if tc.ThinkingLevel != tt.wantLevel {
				t.Errorf("ThinkingLevel = %q, want %q", tc.ThinkingLevel, tt.wantLevel)
			}
			switch {
			case tt.wantBudget == nil && tc.ThinkingBudget != nil:
				t.Errorf("ThinkingBudget = %d, want nil", *tc.ThinkingBudget)
			case tt.wantBudget != nil && (tc.ThinkingBudget == nil || *tc.ThinkingBudget != *tt.wantBudget):
				t.Errorf("ThinkingBudget = %v, want %d", tc.ThinkingBudget, *tt.wantBudget)
			}
		})
	}
}

func TestBuildConfig_Google_ReasoningKeepsSampling(t *testing.T) {
	req := reasoningReq(provider.VertexAI, true, "medium")
	req.TemperatureEnabled = true
	req.Temperature = 0.4

	cfg, err := buildConfig(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gc := cfg.(*genai.GenerateContentConfig)
	if gc.Temperature == nil || *gc.Temperature != float32(0.4) {
		t.Errorf("Temperature = %v, want 0.4", gc.Temperature)
	}
	if gc.ThinkingConfig == nil {
		t.Error("ThinkingConfig is nil")
	}
}

func TestBuildConfig_Anthropic_Reasoning(t *testing.T) {
	req := reasoningReq(provider.Anthropic, true, "2048")
	req.TemperatureEnabled = true
	req.Temperature = 0.7
	req.TopPEnabled = true
	req.TopP = 0.9
	req.TopKEnabled = true
	req.TopK = 40

	cfg, err := buildConfig(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	params := cfg.(*anthropicsdk.MessageNewParams)

	if params.Thinking.OfEnabled == nil {
		t.Fatal("thinking is not enabled")
	}
	if got := params.Thinking.OfEnabled.BudgetTokens; got != 2048 {
		t.Errorf("BudgetTokens = %d, want 2048", got)
	}
	if params.MaxTokens != 2048+4096 {
		t.Errorf("MaxTokens = %d, want %d", params.MaxTokens, 2048+4096)
	}
	if params.MaxTokens <= params.Thinking.OfEnabled.BudgetTokens {
		t.Error("MaxTokens must be greater than the thinking budget")
	}
	if params.Temperature.Valid() || params.TopP.Valid() || params.TopK.Valid() {
		t.Error("sampling options must not be sent while thinking is enabled")
	}
}

func TestBuildConfig_Anthropic_ReasoningDisabledKeepsSampling(t *testing.T) {
	req := reasoningReq(provider.Anthropic, false, "")
	req.TemperatureEnabled = true
	req.Temperature = 0.7
	req.TopKEnabled = true
	req.TopK = 40

	cfg, err := buildConfig(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	params := cfg.(*anthropicsdk.MessageNewParams)
	if !params.Temperature.Valid() || !params.TopK.Valid() {
		t.Error("sampling options must be sent when thinking is disabled")
	}
}

func TestBuildConfig_Anthropic_InvalidReasoning(t *testing.T) {
	for _, value := range []string{"", "high", "abc", "1.5", "1023", "0", "-5"} {
		t.Run(value, func(t *testing.T) {
			cfg, err := buildConfig(reasoningReq(provider.Anthropic, true, value))
			if err == nil {
				t.Fatalf("expected an error, got config %#v", cfg)
			}
			if !strings.Contains(err.Error(), "Anthropic") {
				t.Errorf("error should name the provider: %v", err)
			}
		})
	}
}

func TestBuildConfig_Ollama_Reasoning(t *testing.T) {
	tests := []struct {
		value    string
		wantJSON string
	}{
		{value: "high", wantJSON: `"high"`},
		{value: " Medium ", wantJSON: `"medium"`},
		{value: "true", wantJSON: `true`},
		{value: "FALSE", wantJSON: `false`},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			cfg, err := buildConfig(reasoningReq(provider.Ollama, true, tt.value))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			oc, ok := cfg.(*ollama.GenerateContentConfig)
			if !ok {
				t.Fatalf("expected *ollama.GenerateContentConfig, got %T", cfg)
			}
			if oc.Think == nil {
				t.Fatal("Think is nil")
			}
			got, err := json.Marshal(oc.Think)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(got) != tt.wantJSON {
				t.Errorf("think = %s, want %s", got, tt.wantJSON)
			}
		})
	}
}

func TestBuildConfig_OpenAICompatible_Reasoning(t *testing.T) {
	for _, pt := range []provider.ProviderType{
		provider.OpenAI, provider.AzureOpenAI, provider.XAI, provider.DeepSeek, provider.OpenAICompatible,
	} {
		t.Run(string(pt), func(t *testing.T) {
			cfg, err := buildConfig(reasoningReq(pt, true, " medium "))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			params, ok := cfg.(*openai.ChatCompletionNewParams)
			if !ok {
				t.Fatalf("expected *ChatCompletionNewParams, got %T", cfg)
			}
			if params.ReasoningEffort != openai.ReasoningEffortMedium {
				t.Errorf("ReasoningEffort = %q, want %q", params.ReasoningEffort, openai.ReasoningEffortMedium)
			}
		})
	}
}

func TestBuildConfig_OpenAICompatible_CustomValuePassesThrough(t *testing.T) {
	cfg, err := buildConfig(reasoningReq(provider.OpenAI, true, "xhigh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	params := cfg.(*openai.ChatCompletionNewParams)
	if string(params.ReasoningEffort) != "xhigh" {
		t.Errorf("ReasoningEffort = %q, want xhigh", params.ReasoningEffort)
	}
}

func TestBuildConfig_UnsupportedProvidersIgnoreReasoning(t *testing.T) {
	for _, pt := range []provider.ProviderType{provider.AzureAIFoundry, provider.Bedrock} {
		t.Run(string(pt), func(t *testing.T) {
			cfg, err := buildConfig(reasoningReq(pt, true, "high"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg != nil {
				t.Errorf("expected nil config, got %#v", cfg)
			}
		})
	}
}
