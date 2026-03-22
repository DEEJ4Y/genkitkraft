package httpchatprovider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
)

// Compile-time check that ChatProvider implements the port interface.
var _ chatprovider.ChatProvider = (*ChatProvider)(nil)

// ChatProvider implements chatprovider.ChatProvider using direct HTTP calls.
type ChatProvider struct {
	client *http.Client
}

// NewChatProvider creates a new HTTP-based chat provider.
func NewChatProvider() *ChatProvider {
	return &ChatProvider{
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (cp *ChatProvider) ChatStream(ctx context.Context, req chatprovider.ChatRequest) (<-chan string, <-chan error) {
	tokenCh := make(chan string, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(tokenCh)
		defer close(errCh)

		pt := provider.ProviderType(req.ProviderType)
		var err error

		switch pt {
		case provider.OpenAI, provider.XAI, provider.DeepSeek, provider.Ollama, provider.OpenAICompatible:
			err = cp.streamOpenAICompatible(ctx, req, tokenCh)
		case provider.AzureOpenAI, provider.AzureAIFoundry:
			err = cp.streamAzureOpenAI(ctx, req, tokenCh)
		case provider.Anthropic:
			err = cp.streamAnthropic(ctx, req, tokenCh)
		case provider.GoogleAI:
			err = cp.streamGoogleAI(ctx, req, tokenCh)
		case provider.VertexAI:
			err = fmt.Errorf("vertex AI chat streaming not yet supported (requires service account auth)")
		case provider.Bedrock:
			err = fmt.Errorf("AWS Bedrock chat streaming not yet supported (requires AWS credential chain auth)")
		default:
			err = fmt.Errorf("unsupported provider type: %s", req.ProviderType)
		}

		if err != nil {
			errCh <- err
		}
	}()

	return tokenCh, errCh
}

// openAIChatRequest is the request body for OpenAI-compatible chat completions.
type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	Stream      bool                `json:"stream"`
	Temperature *float64            `json:"temperature,omitempty"`
	TopP        *float64            `json:"top_p,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (cp *ChatProvider) streamOpenAICompatible(ctx context.Context, req chatprovider.ChatRequest, tokenCh chan<- string) error {
	baseURL := defaultBaseURL(provider.ProviderType(req.ProviderType), req.BaseURL)
	url := baseURL + "/v1/chat/completions"

	// Ollama uses /api/chat for its native API but also supports /v1/chat/completions
	body := cp.buildOpenAIBody(req)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, jsonReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if req.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	}

	// Apply custom headers from OpenAI-compatible config
	cp.applyCustomHeaders(httpReq, req)

	return cp.doStreamOpenAI(httpReq, tokenCh)
}

func (cp *ChatProvider) streamAzureOpenAI(ctx context.Context, req chatprovider.ChatRequest, tokenCh chan<- string) error {
	if req.BaseURL == "" {
		return fmt.Errorf("Azure OpenAI requires a base URL")
	}

	// Azure uses a different URL pattern: {base}/openai/deployments/{model}/chat/completions?api-version=...
	// However, some setups use /openai/chat/completions with model in body.
	// We'll use the deployment-based URL if the model looks like a deployment name.
	apiVersion := "2024-10-21"
	if cfg := parseAzureConfig(req.Config); cfg.APIVersion != "" {
		apiVersion = cfg.APIVersion
	}

	var url string
	if cfg := parseAzureConfig(req.Config); cfg.DeploymentName != "" {
		url = fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", req.BaseURL, cfg.DeploymentName, apiVersion)
	} else {
		url = fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", req.BaseURL, req.ModelID, apiVersion)
	}

	body := cp.buildOpenAIBody(req)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, jsonReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key", req.APIKey)

	return cp.doStreamOpenAI(httpReq, tokenCh)
}

func (cp *ChatProvider) buildOpenAIBody(req chatprovider.ChatRequest) openAIChatRequest {
	messages := make([]openAIChatMessage, 0, len(req.Messages)+1)

	// Add system prompt as first message if present
	if req.SystemPrompt != "" {
		messages = append(messages, openAIChatMessage{Role: "system", Content: req.SystemPrompt})
	}

	for _, m := range req.Messages {
		messages = append(messages, openAIChatMessage{Role: m.Role, Content: m.Content})
	}

	body := openAIChatRequest{
		Model:    req.ModelID,
		Messages: messages,
		Stream:   true,
	}
	if req.Temperature > 0 {
		body.Temperature = &req.Temperature
	}
	if req.TopP > 0 {
		body.TopP = &req.TopP
	}
	return body
}

func (cp *ChatProvider) doStreamOpenAI(httpReq *http.Request, tokenCh chan<- string) error {
	resp, err := cp.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("provider returned status %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			select {
			case tokenCh <- chunk.Choices[0].Delta.Content:
			case <-httpReq.Context().Done():
				return httpReq.Context().Err()
			}
		}
	}
	return scanner.Err()
}

type openAIStreamChunk struct {
	Choices []openAIStreamChoice `json:"choices"`
}

type openAIStreamChoice struct {
	Delta openAIStreamDelta `json:"delta"`
}

type openAIStreamDelta struct {
	Content string `json:"content"`
}

// --- Anthropic ---

type anthropicRequest struct {
	Model       string                `json:"model"`
	MaxTokens   int                   `json:"max_tokens"`
	System      string                `json:"system,omitempty"`
	Messages    []anthropicMessage    `json:"messages"`
	Stream      bool                  `json:"stream"`
	Temperature *float64              `json:"temperature,omitempty"`
	TopP        *float64              `json:"top_p,omitempty"`
	TopK        *int                  `json:"top_k,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (cp *ChatProvider) streamAnthropic(ctx context.Context, req chatprovider.ChatRequest, tokenCh chan<- string) error {
	url := "https://api.anthropic.com/v1/messages"

	messages := make([]anthropicMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == "system" {
			continue // system prompt is a top-level field in Anthropic API
		}
		messages = append(messages, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	body := anthropicRequest{
		Model:     req.ModelID,
		MaxTokens: 4096,
		System:    req.SystemPrompt,
		Messages:  messages,
		Stream:    true,
	}
	// Anthropic does not allow both temperature and top_p in the same request.
	// When both are set, prefer temperature and drop top_p.
	if req.Temperature > 0 {
		body.Temperature = &req.Temperature
	} else if req.TopP > 0 {
		body.TopP = &req.TopP
	}
	if req.TopK > 0 {
		body.TopK = &req.TopK
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, jsonReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", req.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := cp.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Anthropic returned status %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event anthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			select {
			case tokenCh <- event.Delta.Text:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		if event.Type == "message_stop" {
			break
		}
	}
	return scanner.Err()
}

type anthropicStreamEvent struct {
	Type  string               `json:"type"`
	Delta anthropicStreamDelta `json:"delta,omitempty"`
}

type anthropicStreamDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// --- Google AI ---

type googleAIRequest struct {
	Contents         []googleAIContent        `json:"contents"`
	SystemInstruction *googleAIContent        `json:"systemInstruction,omitempty"`
	GenerationConfig *googleAIGenerationConfig `json:"generationConfig,omitempty"`
}

type googleAIContent struct {
	Role  string         `json:"role,omitempty"`
	Parts []googleAIPart `json:"parts"`
}

type googleAIPart struct {
	Text string `json:"text"`
}

type googleAIGenerationConfig struct {
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"topP,omitempty"`
	TopK        *int     `json:"topK,omitempty"`
}

func (cp *ChatProvider) streamGoogleAI(ctx context.Context, req chatprovider.ChatRequest, tokenCh chan<- string) error {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s",
		req.ModelID, req.APIKey)

	contents := make([]googleAIContent, 0, len(req.Messages))
	for _, m := range req.Messages {
		role := m.Role
		if role == "assistant" {
			role = "model" // Google AI uses "model" instead of "assistant"
		}
		contents = append(contents, googleAIContent{
			Role:  role,
			Parts: []googleAIPart{{Text: m.Content}},
		})
	}

	body := googleAIRequest{Contents: contents}

	if req.SystemPrompt != "" {
		body.SystemInstruction = &googleAIContent{
			Parts: []googleAIPart{{Text: req.SystemPrompt}},
		}
	}

	genConfig := &googleAIGenerationConfig{}
	hasConfig := false
	if req.Temperature > 0 {
		genConfig.Temperature = &req.Temperature
		hasConfig = true
	}
	if req.TopP > 0 {
		genConfig.TopP = &req.TopP
		hasConfig = true
	}
	if req.TopK > 0 {
		genConfig.TopK = &req.TopK
		hasConfig = true
	}
	if hasConfig {
		body.GenerationConfig = genConfig
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, jsonReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := cp.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Google AI returned status %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var chunk googleAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		for _, candidate := range chunk.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					select {
					case tokenCh <- part.Text:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
		}
	}
	return scanner.Err()
}

type googleAIStreamChunk struct {
	Candidates []googleAICandidate `json:"candidates"`
}

type googleAICandidate struct {
	Content googleAIContent `json:"content"`
}

// --- Helpers ---

func defaultBaseURL(pt provider.ProviderType, baseURL string) string {
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/")
	}
	switch pt {
	case provider.OpenAI:
		return "https://api.openai.com"
	case provider.XAI:
		return "https://api.x.ai"
	case provider.DeepSeek:
		return "https://api.deepseek.com"
	case provider.Ollama:
		return "http://localhost:11434"
	default:
		return ""
	}
}

func jsonReader(v any) io.Reader {
	data, _ := json.Marshal(v)
	return bytes.NewReader(data)
}

func (cp *ChatProvider) applyCustomHeaders(httpReq *http.Request, req chatprovider.ChatRequest) {
	if len(req.Config) == 0 {
		return
	}
	var cfg struct {
		Organization  string `json:"organization"`
		CustomHeaders string `json:"custom_headers"`
	}
	if err := json.Unmarshal(req.Config, &cfg); err != nil {
		return
	}
	if cfg.Organization != "" {
		httpReq.Header.Set("OpenAI-Organization", cfg.Organization)
	}
	if cfg.CustomHeaders != "" {
		for _, line := range strings.Split(cfg.CustomHeaders, "\n") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				httpReq.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}
}

type azureConfig struct {
	DeploymentName string `json:"deployment_name"`
	APIVersion     string `json:"api_version"`
}

func parseAzureConfig(raw json.RawMessage) azureConfig {
	var cfg azureConfig
	if len(raw) > 0 {
		json.Unmarshal(raw, &cfg)
	}
	return cfg
}
