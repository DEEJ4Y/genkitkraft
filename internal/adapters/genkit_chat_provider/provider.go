package genkitchatprovider

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
	"google.golang.org/genai"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
)

// Compile-time check that ChatProvider implements the port interface.
var _ chatprovider.ChatProvider = (*ChatProvider)(nil)

// ChatProvider implements chatprovider.ChatProvider using the Genkit Go SDK.
type ChatProvider struct{}

// NewChatProvider creates a new Genkit-based chat provider.
func NewChatProvider() *ChatProvider {
	return &ChatProvider{}
}

func (cp *ChatProvider) ChatStream(ctx context.Context, req chatprovider.ChatRequest) (<-chan string, <-chan error) {
	tokenCh := make(chan string, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(tokenCh)
		defer close(errCh)

		if err := cp.doStream(ctx, req, tokenCh); err != nil {
			errCh <- err
		}
	}()

	return tokenCh, errCh
}

func (cp *ChatProvider) doStream(ctx context.Context, req chatprovider.ChatRequest, tokenCh chan<- string) error {
	result, err := buildPlugin(req)
	if err != nil {
		return err
	}

	g := genkit.Init(ctx, genkit.WithPlugins(result.plugin))

	// Some plugins require explicit model registration after init.
	if result.postInit != nil {
		result.postInit(g)
	}

	// Build generate options.
	opts := []ai.GenerateOption{
		ai.WithModelName(result.modelName),
	}

	if req.SystemPrompt != "" {
		opts = append(opts, ai.WithSystem(req.SystemPrompt))
	}

	// Convert conversation history to Genkit messages.
	if len(req.Messages) > 0 {
		messages := buildMessages(req.Messages)
		opts = append(opts, ai.WithMessages(messages...))
	}

	// Build provider-specific config.
	if cfg := buildConfig(req); cfg != nil {
		opts = append(opts, ai.WithConfig(cfg))
	}

	// Stream the response.
	stream := genkit.GenerateStream(ctx, g, opts...)
	for chunk, err := range stream {
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}
		if chunk.Done {
			break
		}
		text := chunk.Chunk.Text()
		if text != "" {
			select {
			case tokenCh <- text:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

// buildMessages converts ChatMessage slice to Genkit Message slice.
func buildMessages(messages []chatprovider.ChatMessage) []*ai.Message {
	result := make([]*ai.Message, 0, len(messages))
	for _, m := range messages {
		switch m.Role {
		case "user":
			result = append(result, ai.NewUserTextMessage(m.Content))
		case "assistant":
			result = append(result, ai.NewModelTextMessage(m.Content))
		case "system":
			result = append(result, ai.NewSystemTextMessage(m.Content))
		}
	}
	return result
}

// buildConfig creates the provider-specific generation config based on the provider type.
func buildConfig(req chatprovider.ChatRequest) any {
	pt := provider.ProviderType(req.ProviderType)

	switch pt {
	case provider.GoogleAI, provider.VertexAI:
		return buildGoogleConfig(req)
	case provider.Anthropic:
		return buildAnthropicConfig(req)
	case provider.Ollama:
		return buildOllamaConfig(req)
	default:
		// OpenAI and compatible providers: genkit handles default config.
		// Only pass config if we have parameters to set.
		return buildOpenAICompatibleConfig(req)
	}
}

func buildGoogleConfig(req chatprovider.ChatRequest) *genai.GenerateContentConfig {
	cfg := &genai.GenerateContentConfig{}
	hasConfig := false

	if req.Temperature > 0 {
		cfg.Temperature = genai.Ptr(float32(req.Temperature))
		hasConfig = true
	}
	if req.TopP > 0 {
		cfg.TopP = genai.Ptr(float32(req.TopP))
		hasConfig = true
	}
	if req.TopK > 0 {
		cfg.TopK = genai.Ptr(float32(req.TopK))
		hasConfig = true
	}

	if !hasConfig {
		return nil
	}
	return cfg
}

func buildAnthropicConfig(req chatprovider.ChatRequest) *anthropicsdk.MessageNewParams {
	cfg := &anthropicsdk.MessageNewParams{
		MaxTokens: 4096,
	}
	hasConfig := true // Always set max_tokens for Anthropic.

	if req.Temperature > 0 {
		cfg.Temperature = anthropicsdk.Float(req.Temperature)
		hasConfig = true
	}
	if req.TopP > 0 {
		cfg.TopP = anthropicsdk.Float(req.TopP)
		hasConfig = true
	}
	if req.TopK > 0 {
		cfg.TopK = anthropicsdk.Int(int64(req.TopK))
		hasConfig = true
	}

	if !hasConfig {
		return nil
	}
	return cfg
}

func buildOllamaConfig(req chatprovider.ChatRequest) *ollama.GenerateContentConfig {
	cfg := &ollama.GenerateContentConfig{}
	hasConfig := false

	if req.Temperature > 0 {
		cfg.Temperature = &req.Temperature
		hasConfig = true
	}
	if req.TopP > 0 {
		cfg.TopP = &req.TopP
		hasConfig = true
	}
	if req.TopK > 0 {
		cfg.TopK = &req.TopK
		hasConfig = true
	}

	if !hasConfig {
		return nil
	}
	return cfg
}

// buildOpenAICompatibleConfig returns nil since OpenAI-compatible providers
// handle temperature/topP/topK through the default genkit pipeline.
// The genkit compat_oai plugin maps these from the standard generate options.
func buildOpenAICompatibleConfig(req chatprovider.ChatRequest) any {
	// The OpenAI-compatible plugins in genkit don't require explicit config
	// for basic temperature/topP parameters — they are passed through the
	// standard generate request. Return nil to use defaults.
	return nil
}
