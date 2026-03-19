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
	default:
		return false, fmt.Sprintf("unknown provider type: %s", p.ProviderType), nil
	}
}

func (t *Tester) testAnthropic(ctx context.Context, p *provider.Provider) (bool, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		return false, "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("x-api-key", p.APIKey)
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
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	return t.doTest(req, "OpenAI")
}

func (t *Tester) testGoogleAI(ctx context.Context, p *provider.Provider) (bool, string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models?key=%s", p.APIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, "", fmt.Errorf("creating request: %w", err)
	}

	return t.doTest(req, "Google AI")
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
