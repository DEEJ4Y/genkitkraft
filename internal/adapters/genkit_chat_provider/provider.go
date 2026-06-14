package genkitchatprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/mcp"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/openai/openai-go"
	"google.golang.org/genai"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
)

// Compile-time check that ChatProvider implements the port interface.
var _ chatprovider.ChatProvider = (*ChatProvider)(nil)

// ChatProvider implements chatprovider.ChatProvider using the Genkit Go SDK.
type ChatProvider struct {
	cache cache.Cache
}

// NewChatProvider creates a new Genkit-based chat provider.
func NewChatProvider(c cache.Cache) *ChatProvider {
	return &ChatProvider{cache: c}
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

func (cp *ChatProvider) Chat(ctx context.Context, req chatprovider.ChatRequest) (string, error) {
	result, err := buildPlugin(req)
	if err != nil {
		return "", err
	}

	g := genkit.Init(ctx, genkit.WithPlugins(result.plugin))
	model := result.getModel(g)

	opts := []ai.GenerateOption{
		ai.WithModel(model),
	}

	if req.SystemPrompt != "" {
		opts = append(opts, ai.WithSystem(req.SystemPrompt))
	}

	if len(req.Messages) > 0 {
		messages := buildMessages(req.Messages)
		opts = append(opts, ai.WithMessages(messages...))
	}

	if cfg := buildConfig(req); cfg != nil {
		opts = append(opts, ai.WithConfig(cfg))
	}

	// Build and attach tools
	tools, mcpCleanup, err := cp.buildTools(ctx, req)
	if err != nil {
		return "", fmt.Errorf("building tools: %w", err)
	}
	if mcpCleanup != nil {
		defer mcpCleanup()
	}
	if len(tools) > 0 {
		opts = append(opts, ai.WithTools(tools...))
	}

	if req.MaxToolCalls > 0 {
		opts = append(opts, ai.WithMaxTurns(req.MaxToolCalls))
	}

	resp, err := genkit.Generate(ctx, g, opts...)
	if err != nil {
		return "", fmt.Errorf("generate error: %w", err)
	}

	return resp.Text(), nil
}

func (cp *ChatProvider) doStream(ctx context.Context, req chatprovider.ChatRequest, tokenCh chan<- string) error {
	result, err := buildPlugin(req)
	if err != nil {
		return err
	}

	g := genkit.Init(ctx, genkit.WithPlugins(result.plugin))

	model := result.getModel(g)

	// Build generate options.
	opts := []ai.GenerateOption{
		ai.WithModel(model),
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

	// Build and attach tools
	tools, mcpCleanup, err := cp.buildTools(ctx, req)
	if err != nil {
		return fmt.Errorf("building tools: %w", err)
	}
	if mcpCleanup != nil {
		defer mcpCleanup()
	}
	if len(tools) > 0 {
		opts = append(opts, ai.WithTools(tools...))
	}

	if req.MaxToolCalls > 0 {
		opts = append(opts, ai.WithMaxTurns(req.MaxToolCalls))
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

	if req.TemperatureEnabled {
		cfg.Temperature = genai.Ptr(float32(req.Temperature))
		hasConfig = true
	}
	if req.TopPEnabled {
		cfg.TopP = genai.Ptr(float32(req.TopP))
		hasConfig = true
	}
	if req.TopKEnabled {
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

	if req.TemperatureEnabled {
		cfg.Temperature = anthropicsdk.Float(req.Temperature)
		hasConfig = true
	}
	if req.TopPEnabled {
		cfg.TopP = anthropicsdk.Float(req.TopP)
		hasConfig = true
	}
	if req.TopKEnabled {
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

	if req.TemperatureEnabled {
		cfg.Temperature = &req.Temperature
		hasConfig = true
	}
	if req.TopPEnabled {
		cfg.TopP = &req.TopP
		hasConfig = true
	}
	if req.TopKEnabled {
		cfg.TopK = &req.TopK
		hasConfig = true
	}

	if !hasConfig {
		return nil
	}
	return cfg
}

// buildOpenAICompatibleConfig builds an openai.ChatCompletionNewParams with
// temperature and topP when enabled. The OpenAI API does not support topK,
// so that parameter is intentionally skipped.
func buildOpenAICompatibleConfig(req chatprovider.ChatRequest) any {
	cfg := &openai.ChatCompletionNewParams{}
	hasConfig := false

	if req.TemperatureEnabled {
		cfg.Temperature = openai.Float(req.Temperature)
		hasConfig = true
	}
	if req.TopPEnabled {
		cfg.TopP = openai.Float(req.TopP)
		hasConfig = true
	}
	// Note: topK is not supported by the OpenAI API and is intentionally skipped.

	if !hasConfig {
		return nil
	}
	return cfg
}

// buildTools creates Genkit tool references from the ChatRequest's HTTP tools and MCP server configs.
// Returns a cleanup function that should be deferred to close MCP connections.
func (cp *ChatProvider) buildTools(ctx context.Context, req chatprovider.ChatRequest) ([]ai.ToolRef, func(), error) {
	var tools []ai.ToolRef
	var mcpClients []*mcp.GenkitMCPClient

	cleanup := func() {
		for _, c := range mcpClients {
			_ = c.Disconnect()
		}
	}

	// Build built-in tools
	builtInTools := cp.buildBuiltInTools(req.BuiltInToolIDs)
	tools = append(tools, builtInTools...)

	// Build HTTP tools
	for _, ht := range req.HttpTools {
		tool := buildHttpTool(ht)
		tools = append(tools, tool)
	}

	// Build MCP tools
	for _, mc := range req.McpServers {
		mcpTools, client, err := buildMcpTools(ctx, mc)
		if err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("MCP server %s: %w", mc.ServerID, err)
		}
		if client != nil {
			mcpClients = append(mcpClients, client)
		}
		tools = append(tools, mcpTools...)
	}

	if len(mcpClients) == 0 {
		return tools, nil, nil
	}
	return tools, cleanup, nil
}

// buildHttpTool creates a Genkit tool from an HTTP tool definition.
func buildHttpTool(ht chatprovider.HttpToolDefinition) ai.Tool {
	name := fmt.Sprintf("http_%s", ht.ID)
	description := ht.Description
	if description == "" {
		description = fmt.Sprintf("HTTP %s %s", ht.Method, ht.URL)
	}

	var inputSchema map[string]any
	if len(ht.InputSchema) > 0 {
		_ = json.Unmarshal(ht.InputSchema, &inputSchema)
	}
	if inputSchema == nil {
		inputSchema = map[string]any{}
	}
	inputSchema = sanitizeSchema(inputSchema)

	toolFn := func(toolCtx *ai.ToolContext, args any) (any, error) {
		return executeHttpTool(toolCtx.Context, ht, args)
	}

	if len(inputSchema) > 0 {
		return ai.NewTool(name, description, toolFn, ai.WithInputSchema(inputSchema))
	}
	return ai.NewTool(name, description, toolFn)
}

// executeHttpTool performs the HTTP request for an HTTP tool invocation.
func executeHttpTool(ctx context.Context, ht chatprovider.HttpToolDefinition, args any) (any, error) {
	// Convert args to map for template rendering
	var argsMap map[string]any
	if args != nil {
		b, err := json.Marshal(args)
		if err != nil {
			return nil, fmt.Errorf("marshaling tool args: %w", err)
		}
		if err := json.Unmarshal(b, &argsMap); err != nil {
			argsMap = map[string]any{}
		}
	}
	if argsMap == nil {
		argsMap = map[string]any{}
	}

	// Render URL template
	resolvedURL, err := renderTemplate("url", ht.URL, argsMap)
	if err != nil {
		return nil, fmt.Errorf("rendering URL template: %w", err)
	}

	// Encode URL path segments and query parameters
	resolvedURL, err = sanitizeURL(resolvedURL)
	if err != nil {
		return nil, fmt.Errorf("invalid resolved URL: %w", err)
	}

	// Render body template
	var bodyReader io.Reader
	if ht.BodyTemplate != "" {
		bodyStr, err := renderTemplate("body", ht.BodyTemplate, argsMap)
		if err != nil {
			return nil, fmt.Errorf("rendering body template: %w", err)
		}
		bodyReader = strings.NewReader(bodyStr)
	}

	httpReq, err := http.NewRequestWithContext(ctx, ht.Method, resolvedURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	// Set headers
	for k, v := range ht.Headers {
		httpReq.Header.Set(k, v)
	}
	if ht.BodyTemplate != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP tool request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("reading HTTP tool response: %w", err)
	}

	// Try to parse as JSON; fall back to string
	var result any
	if err := json.Unmarshal(respBody, &result); err != nil {
		result = string(respBody)
	}

	return map[string]any{
		"status": resp.StatusCode,
		"body":   result,
	}, nil
}

// renderTemplate renders a Go text/template with the provided data.
func renderTemplate(name, tmplStr string, data map[string]any) (string, error) {
	tmpl, err := template.New(name).Parse(tmplStr)
	if err != nil {
		return tmplStr, nil // If template is invalid, use as-is
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return tmplStr, nil
	}
	return buf.String(), nil
}

// buildMcpTools connects to an MCP server and returns the selected tools.
// It tries the configured transport first and falls back to the alternate
// transport (SSE ↔ Streamable HTTP) on failure.
// Tool input schemas are sanitized to remove anyOf/oneOf/allOf constructs
// that some providers (e.g. Anthropic) do not support.
func buildMcpTools(ctx context.Context, mc chatprovider.McpServerConfig) ([]ai.ToolRef, *mcp.GenkitMCPClient, error) {
	client, err := connectMcpWithFallback(mc)
	if err != nil {
		return nil, nil, err
	}

	allTools, err := client.GetActiveTools(ctx, nil)
	if err != nil {
		_ = client.Disconnect()
		return nil, nil, fmt.Errorf("listing MCP tools: %w", err)
	}

	// Build selected tool name set for filtering
	var selectedSet map[string]bool
	if !mc.SelectAll {
		selectedSet = make(map[string]bool, len(mc.ToolNames))
		for _, name := range mc.ToolNames {
			selectedSet[client.GetToolNameWithNamespace(name)] = true
		}
	}

	var refs []ai.ToolRef
	for _, origTool := range allTools {
		if selectedSet != nil && !selectedSet[origTool.Name()] {
			continue
		}
		wrapped := wrapToolWithSanitizedSchema(origTool)
		refs = append(refs, wrapped)
	}

	return refs, client, nil
}

// wrapToolWithSanitizedSchema creates a new ai.Tool that delegates execution
// to the original tool but uses a sanitized input schema (strips anyOf/oneOf/allOf).
func wrapToolWithSanitizedSchema(orig ai.Tool) ai.Tool {
	def := orig.Definition()
	sanitized := sanitizeSchema(def.InputSchema)

	fn := func(toolCtx *ai.ToolContext, args any) (any, error) {
		return orig.RunRaw(toolCtx.Context, args)
	}

	if len(sanitized) > 0 {
		return ai.NewTool(def.Name, def.Description, fn, ai.WithInputSchema(sanitized))
	}
	return ai.NewTool(def.Name, def.Description, fn)
}

// sanitizeSchema recursively removes anyOf, oneOf, and allOf from a JSON Schema,
// replacing them with the first variant. These constructs are not supported by
// all LLM providers (notably Anthropic).
func sanitizeSchema(schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}

	result := make(map[string]any, len(schema))
	for k, v := range schema {
		switch k {
		case "anyOf", "oneOf":
			// Replace with the first variant's properties
			if variants, ok := v.([]any); ok && len(variants) > 0 {
				if first, ok := variants[0].(map[string]any); ok {
					for fk, fv := range sanitizeSchema(first) {
						if _, exists := result[fk]; !exists {
							result[fk] = fv
						}
					}
				}
			}
		case "allOf":
			// Merge all variants
			if variants, ok := v.([]any); ok {
				for _, variant := range variants {
					if m, ok := variant.(map[string]any); ok {
						for fk, fv := range sanitizeSchema(m) {
							if _, exists := result[fk]; !exists {
								result[fk] = fv
							}
						}
					}
				}
			}
		default:
			// Recursively sanitize nested schemas
			switch val := v.(type) {
			case map[string]any:
				result[k] = sanitizeSchema(val)
			case []any:
				arr := make([]any, len(val))
				for i, item := range val {
					if m, ok := item.(map[string]any); ok {
						arr[i] = sanitizeSchema(m)
					} else {
						arr[i] = item
					}
				}
				result[k] = arr
			default:
				result[k] = v
			}
		}
	}
	return result
}

// connectMcpWithFallback tries the configured transport, then falls back to
// the alternate (SSE ↔ Streamable HTTP) if the first attempt fails.
func connectMcpWithFallback(mc chatprovider.McpServerConfig) (*mcp.GenkitMCPClient, error) {
	primary, fallback := mcpTransportOrder(mc.Transport)

	client, err := connectMcp(mc.ServerID, primary, mc.URL, mc.Headers)
	if err == nil {
		return client, nil
	}

	client, fallbackErr := connectMcp(mc.ServerID, fallback, mc.URL, mc.Headers)
	if fallbackErr == nil {
		return client, nil
	}

	return nil, fmt.Errorf("connecting to MCP server: %s failed: %w; fallback %s also failed: %v", primary, err, fallback, fallbackErr)
}

func connectMcp(serverID, transport, url string, headers map[string]string) (*mcp.GenkitMCPClient, error) {
	opts := mcp.MCPClientOptions{
		Name: fmt.Sprintf("mcp_%s", serverID),
	}
	switch transport {
	case "sse":
		opts.SSE = &mcp.SSEConfig{BaseURL: url, Headers: headers}
	case "streamableHttp":
		opts.StreamableHTTP = &mcp.StreamableHTTPConfig{BaseURL: url, Headers: headers}
	default:
		return nil, fmt.Errorf("unsupported transport: %s", transport)
	}
	return mcp.NewGenkitMCPClient(opts)
}

func mcpTransportOrder(configured string) (string, string) {
	if strings.EqualFold(configured, "sse") {
		return "sse", "streamableHttp"
	}
	return "streamableHttp", "sse"
}
