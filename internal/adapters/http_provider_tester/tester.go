package httpprovidertester

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	providertester "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_tester"
)

// Compile-time check that Tester implements the port interface.
var _ providertester.Tester = (*Tester)(nil)

// Tester tests provider API connectivity via HTTP.
type Tester struct {
	client *http.Client
}

// NewTester creates a new HTTP-based provider tester.
func NewTester() *Tester {
	return &Tester{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *Tester) Test(ctx context.Context, p *provider.Provider) (bool, string, error) {
	switch p.ProviderType {
	case provider.Anthropic:
		return t.testAnthropic(ctx, p)
	case provider.OpenAI:
		return t.testOpenAI(ctx, p)
	case provider.GoogleAI:
		return t.testGoogleAI(ctx, p)
	case provider.Ollama:
		return t.testOllama(ctx, p)
	case provider.XAI:
		return t.testXAI(ctx, p)
	case provider.DeepSeek:
		return t.testDeepSeek(ctx, p)
	case provider.AzureOpenAI:
		return t.testAzureOpenAI(ctx, p)
	case provider.VertexAI:
		return false, "Vertex AI connectivity test not yet implemented (uses service account auth)", nil
	case provider.Bedrock:
		return false, "AWS Bedrock connectivity test not yet implemented (coming soon)", nil
	default:
		return false, fmt.Sprintf("unknown provider type: %s", p.ProviderType), nil
	}
}

func (t *Tester) apiKey(p *provider.Provider) string {
	if p.APIKey != nil {
		return *p.APIKey
	}
	return ""
}

func (t *Tester) testAnthropic(ctx context.Context, p *provider.Provider) (bool, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		return false, "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("x-api-key", t.apiKey(p))
	req.Header.Set("anthropic-version", "2023-06-01")

	return t.doTest(req, "Anthropic")
}

func (t *Tester) testOpenAI(ctx context.Context, p *provider.Provider) (bool, string, error) {
	baseURL := "https://api.openai.com"
	if p.BaseURL != "" {
		baseURL = p.BaseURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/models", nil)
	if err != nil {
		return false, "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey(p))

	return t.doTest(req, "OpenAI")
}

func (t *Tester) testGoogleAI(ctx context.Context, p *provider.Provider) (bool, string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models?key=%s", t.apiKey(p))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, "", fmt.Errorf("creating request: %w", err)
	}

	return t.doTest(req, "Google AI")
}

func (t *Tester) testOllama(ctx context.Context, p *provider.Provider) (bool, string, error) {
	baseURL := p.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/tags", nil)
	if err != nil {
		return false, "", fmt.Errorf("creating request: %w", err)
	}

	return t.doTest(req, "Ollama")
}

func (t *Tester) testXAI(ctx context.Context, p *provider.Provider) (bool, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.x.ai/v1/models", nil)
	if err != nil {
		return false, "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey(p))

	return t.doTest(req, "xAI")
}

func (t *Tester) testDeepSeek(ctx context.Context, p *provider.Provider) (bool, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.deepseek.com/v1/models", nil)
	if err != nil {
		return false, "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey(p))

	return t.doTest(req, "DeepSeek")
}

func (t *Tester) testAzureOpenAI(ctx context.Context, p *provider.Provider) (bool, string, error) {
	if p.BaseURL == "" {
		return false, "Azure OpenAI requires a base URL (endpoint)", nil
	}

	// Use the Azure models list endpoint
	url := fmt.Sprintf("%s/openai/models?api-version=2024-10-21", p.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("api-key", t.apiKey(p))

	return t.doTest(req, "Azure OpenAI")
}

func (t *Tester) doTest(req *http.Request, providerName string) (bool, string, error) {
	resp, err := t.client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err), nil
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusOK {
		return true, fmt.Sprintf("%s connection successful", providerName), nil
	}

	return false, fmt.Sprintf("%s returned status %d", providerName, resp.StatusCode), nil
}
