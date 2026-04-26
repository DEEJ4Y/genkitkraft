package mcpdiscoveryadapter

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/DEEJ4Y/genkitkraft/internal/ports/mcp_discovery"
	"github.com/firebase/genkit/go/plugins/mcp"
)

var _ mcpdiscovery.McpDiscovery = (*Adapter)(nil)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) ListTools(ctx context.Context, serverName, transport, url string, headers map[string]string) ([]mcpdiscovery.McpTool, error) {
	slug := slugify(serverName)
	client, err := connectWithFallback(ctx, slug, transport, url, headers)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect()

	allTools, err := client.GetActiveTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("listing MCP tools: %w", err)
	}

	tools := make([]mcpdiscovery.McpTool, len(allTools))
	for i, t := range allTools {
		tools[i] = mcpdiscovery.McpTool{
			Name:        t.Name(),
			Description: t.Definition().Description,
		}
	}
	return tools, nil
}

// connectWithFallback tries the configured transport first; on failure it
// falls back to the alternate transport (SSE ↔ Streamable HTTP).
func connectWithFallback(ctx context.Context, slug, transport, url string, headers map[string]string) (*mcp.GenkitMCPClient, error) {
	primary, fallback := transportOrder(transport)

	client, err := connect(ctx, slug, primary, url, headers)
	if err == nil {
		return client, nil
	}

	// Try fallback transport
	client, fallbackErr := connect(ctx, slug, fallback, url, headers)
	if fallbackErr == nil {
		return client, nil
	}

	return nil, fmt.Errorf("failed to connect with %s: %w; fallback %s also failed: %v", primary, err, fallback, fallbackErr)
}

func connect(_ context.Context, slug, transport, url string, headers map[string]string) (*mcp.GenkitMCPClient, error) {
	opts := mcp.MCPClientOptions{
		Name: slug,
	}

	switch transport {
	case "sse":
		opts.SSE = &mcp.SSEConfig{
			BaseURL: url,
			Headers: headers,
		}
	case "streamableHttp":
		opts.StreamableHTTP = &mcp.StreamableHTTPConfig{
			BaseURL: url,
			Headers: headers,
		}
	default:
		return nil, fmt.Errorf("unsupported transport: %s", transport)
	}

	return mcp.NewGenkitMCPClient(opts)
}

// transportOrder returns the primary and fallback transport names.
func transportOrder(configured string) (string, string) {
	if strings.EqualFold(configured, "sse") {
		return "sse", "streamableHttp"
	}
	return "streamableHttp", "sse"
}

// slugify converts a server name to a lowercase slug (e.g. "My Todo Server" → "my-todo-server").
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "mcp-server"
	}
	return s
}
