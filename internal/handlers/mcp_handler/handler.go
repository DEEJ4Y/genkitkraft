package mcphandler

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/config"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
	mcpdiscovery "github.com/DEEJ4Y/genkitkraft/internal/ports/mcp_discovery"
)

// Handler registers all GenKitKraft APIs as MCP tools and serves them
// over the StreamableHTTP transport.
type Handler struct {
	authApp       *app.AuthApp
	providerApp   *app.ProviderApp
	promptApp     *app.PromptApp
	agentApp      *app.AgentApp
	playgroundApp *app.PlaygroundApp
	httpToolApp   *app.HttpToolApp
	mcpServerApp  *app.McpServerApp
	agentToolApp  *app.AgentToolApp
	chatProvider  chatprovider.ChatProvider
	mcpDiscovery  mcpdiscovery.McpDiscovery
	authCfg       config.AuthConfig
}

func NewHandler(
	authApp *app.AuthApp,
	providerApp *app.ProviderApp,
	promptApp *app.PromptApp,
	agentApp *app.AgentApp,
	playgroundApp *app.PlaygroundApp,
	httpToolApp *app.HttpToolApp,
	mcpServerApp *app.McpServerApp,
	agentToolApp *app.AgentToolApp,
	chatProvider chatprovider.ChatProvider,
	mcpDiscovery mcpdiscovery.McpDiscovery,
	authCfg config.AuthConfig,
) *Handler {
	return &Handler{
		authApp:       authApp,
		providerApp:   providerApp,
		promptApp:     promptApp,
		agentApp:      agentApp,
		playgroundApp: playgroundApp,
		httpToolApp:   httpToolApp,
		mcpServerApp:  mcpServerApp,
		agentToolApp:  agentToolApp,
		chatProvider:  chatProvider,
		mcpDiscovery:  mcpDiscovery,
		authCfg:       authCfg,
	}
}

// HTTPHandler builds an MCP server with all registered tools and returns
// an http.Handler ready to be mounted on a ServeMux. If AUTH_CREDENTIALS is
// configured, the handler is wrapped with HTTP basic auth.
func (h *Handler) HTTPHandler() http.Handler {
	server := mcp.NewServer(
		&mcp.Implementation{Name: "genkitkraft", Version: "v1.0.0"},
		nil,
	)

	h.registerAuthTools(server)
	h.registerAgentTools(server)
	h.registerAgentToolTools(server)
	h.registerPromptTools(server)
	h.registerProviderTools(server)
	h.registerHttpToolTools(server)
	h.registerMcpServerTools(server)
	h.registerPlaygroundTools(server)
	h.registerHealthTools(server)

	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)

	if len(h.authCfg.Credentials) > 0 {
		return basicAuthMiddleware(handler, h.authCfg.Credentials)
	}
	return handler
}
