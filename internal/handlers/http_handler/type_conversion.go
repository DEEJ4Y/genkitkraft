package httphandler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	builtintool "github.com/DEEJ4Y/genkitkraft/internal/domain/builtin_tool"
	httptool "github.com/DEEJ4Y/genkitkraft/internal/domain/http_tool"
	mcpserver "github.com/DEEJ4Y/genkitkraft/internal/domain/mcp_server"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	agenttoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_tool_repo"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
)

func toLoginParams(req gen.ModelsLoginRequest, clientIP string) commands.LoginParams {
	return commands.LoginParams{
		Username: req.Username,
		Password: req.Password,
		ClientIP: clientIP,
	}
}

func toLoginResponse(result commands.LoginResult) gen.ModelsLoginResponse {
	return gen.ModelsLoginResponse{Username: result.Username}
}

func toMeResponse(result queries.GetMeResult) gen.ModelsMeResponse {
	return gen.ModelsMeResponse{Username: result.Username}
}

func toAuthStatusResponse(result queries.GetAuthStatusResult) gen.ModelsAuthStatusResponse {
	return gen.ModelsAuthStatusResponse{Required: result.Required}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeAppError(w http.ResponseWriter, err error) {
	if appErr, ok := errors.IsAppError(err); ok {
		writeJSON(w, errors.HTTPStatusCode(appErr.Code()), gen.ModelsErrorResponse{Error: appErr.Error()})
		return
	}
	writeJSON(w, http.StatusInternalServerError, gen.ModelsErrorResponse{Error: "internal server error"})
}

func extractIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.SplitN(fwd, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

func isSecure(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func toProviderResponse(p *provider.Provider) gen.ModelsProviderResponse {
	resp := gen.ModelsProviderResponse{
		Id:           p.ID,
		Name:         p.Name,
		ProviderType: gen.ModelsProviderType(p.ProviderType),
		ApiKey:       p.MaskedAPIKey(),
		Enabled:      p.Enabled,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	if p.BaseURL != "" {
		resp.BaseUrl = &p.BaseURL
	}
	// Include config if non-empty
	if len(p.RawConfig) > 0 && string(p.RawConfig) != "{}" {
		var configMap map[string]string
		if err := json.Unmarshal(p.RawConfig, &configMap); err == nil && len(configMap) > 0 {
			resp.Config = &configMap
		}
	}
	return resp
}

func toProviderListResponse(result queries.ListProvidersResult) gen.ModelsProviderListResponse {
	providers := make([]gen.ModelsProviderResponse, len(result.Providers))
	for i, p := range result.Providers {
		providers[i] = toProviderResponse(p)
	}
	return gen.ModelsProviderListResponse{Providers: providers}
}

func toCreateProviderParams(req gen.ModelsCreateProviderRequest) commands.CreateProviderParams {
	params := commands.CreateProviderParams{
		Name:         req.Name,
		ProviderType: provider.ProviderType(req.ProviderType),
		APIKey:       req.ApiKey,
	}
	if req.BaseUrl != nil {
		params.BaseURL = *req.BaseUrl
	}
	if req.Config != nil {
		params.Config = *req.Config
	}
	return params
}

func toUpdateProviderParams(id string, req gen.ModelsUpdateProviderRequest) commands.UpdateProviderParams {
	return commands.UpdateProviderParams{
		ID:      id,
		Name:    req.Name,
		APIKey:  req.ApiKey,
		BaseURL: req.BaseUrl,
		Config:  req.Config,
		Enabled: req.Enabled,
	}
}

func toPromptResponse(p *prompt.Prompt) gen.ModelsPromptResponse {
	return gen.ModelsPromptResponse{
		Id:        p.ID,
		Name:      p.Name,
		Content:   p.Content,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toPromptListResponse(result queries.ListPromptsResult, limit, offset int) gen.ModelsPromptListResponse {
	prompts := make([]gen.ModelsPromptResponse, len(result.Prompts))
	for i, p := range result.Prompts {
		prompts[i] = toPromptResponse(p)
	}
	return gen.ModelsPromptListResponse{
		Prompts: prompts,
		Total:   int32(result.Total),
		Limit:   int32(limit),
		Offset:  int32(offset),
	}
}

func toCreatePromptParams(req gen.ModelsCreatePromptRequest) commands.CreatePromptParams {
	return commands.CreatePromptParams{
		Name:    req.Name,
		Content: req.Content,
	}
}

func toUpdatePromptParams(id string, req gen.ModelsUpdatePromptRequest) commands.UpdatePromptParams {
	return commands.UpdatePromptParams{
		ID:      id,
		Name:    req.Name,
		Content: req.Content,
	}
}

func toListPromptsParams(params gen.ListPromptsParams) queries.ListPromptsParams {
	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	return queries.ListPromptsParams{
		Limit:  limit,
		Offset: offset,
	}
}

func toProviderTypeListResponse(result queries.ListProviderTypesResult) gen.ModelsProviderTypeListResponse {
	types := make([]gen.ModelsProviderTypeInfo, len(result.ProviderTypes))
	for i, pt := range result.ProviderTypes {
		info := gen.ModelsProviderTypeInfo{
			Type:            gen.ModelsProviderType(pt.Type),
			DisplayName:     pt.DisplayName,
			RequiresApiKey:  pt.RequiresAPIKey,
			RequiresBaseUrl: pt.RequiresBaseURL,
			EnvVarHint:      pt.EnvVarHint,
			ModelPrefix:     pt.ModelPrefix,
		}
		if pt.BaseURLDefault != "" {
			info.BaseUrlDefault = &pt.BaseURLDefault
		}
		if pt.ComingSoon {
			comingSoon := true
			info.ComingSoon = &comingSoon
		}
		fields := make([]gen.ModelsConfigFieldInfo, len(pt.ConfigFields))
		for j, f := range pt.ConfigFields {
			field := gen.ModelsConfigFieldInfo{
				Name:     f.Name,
				Label:    f.Label,
				Required: f.Required,
			}
			if f.Placeholder != "" {
				field.Placeholder = &f.Placeholder
			}
			if f.Sensitive {
				sensitive := true
				field.Sensitive = &sensitive
			}
			fields[j] = field
		}
		info.ConfigFields = fields
		types[i] = info
	}
	return gen.ModelsProviderTypeListResponse{ProviderTypes: types}
}

func toAgentResponse(a *agent.Agent) gen.ModelsAgentResponse {
	resp := gen.ModelsAgentResponse{
		Id:                 a.ID,
		Name:               a.Name,
		ProviderId:         a.ProviderID,
		ProviderName:       a.ProviderName,
		ProviderType:       gen.ModelsProviderType(a.ProviderType),
		ModelId:            a.ModelID,
		TemperatureEnabled: a.TemperatureEnabled,
		Temperature:        float32(a.Temperature),
		TopPEnabled:        a.TopPEnabled,
		TopP:               float32(a.TopP),
		TopKEnabled:        a.TopKEnabled,
		TopK:               int32(a.TopK),
		MaxToolCalls:       int32(a.MaxToolCalls),
		CreatedAt:          a.CreatedAt,
		UpdatedAt:          a.UpdatedAt,
	}
	if a.SystemPromptID != "" {
		resp.SystemPromptId = &a.SystemPromptID
	}
	if a.SystemPromptName != "" {
		resp.SystemPromptName = &a.SystemPromptName
	}
	return resp
}

func toAgentListResponse(result queries.ListAgentsResult, limit, offset int) gen.ModelsAgentListResponse {
	agents := make([]gen.ModelsAgentResponse, len(result.Agents))
	for i, a := range result.Agents {
		agents[i] = toAgentResponse(a)
	}
	return gen.ModelsAgentListResponse{
		Agents: agents,
		Total:  int32(result.Total),
		Limit:  int32(limit),
		Offset: int32(offset),
	}
}

func toCreateAgentParams(req gen.ModelsCreateAgentRequest) commands.CreateAgentParams {
	params := commands.CreateAgentParams{
		Name:       req.Name,
		ProviderID: req.ProviderId,
		ModelID:    req.ModelId,
	}
	if req.SystemPromptId != nil {
		params.SystemPromptID = *req.SystemPromptId
	}
	params.TemperatureEnabled = req.TemperatureEnabled
	if req.Temperature != nil {
		t := float64(*req.Temperature)
		params.Temperature = &t
	}
	params.TopPEnabled = req.TopPEnabled
	if req.TopP != nil {
		t := float64(*req.TopP)
		params.TopP = &t
	}
	params.TopKEnabled = req.TopKEnabled
	if req.TopK != nil {
		t := int(*req.TopK)
		params.TopK = &t
	}
	if req.MaxToolCalls != nil {
		t := int(*req.MaxToolCalls)
		params.MaxToolCalls = &t
	}
	return params
}

func toUpdateAgentParams(id string, req gen.ModelsUpdateAgentRequest) commands.UpdateAgentParams {
	params := commands.UpdateAgentParams{
		ID: id,
	}
	params.Name = req.Name
	params.ProviderID = req.ProviderId
	params.ModelID = req.ModelId
	params.SystemPromptID = req.SystemPromptId
	params.TemperatureEnabled = req.TemperatureEnabled
	if req.Temperature != nil {
		t := float64(*req.Temperature)
		params.Temperature = &t
	}
	params.TopPEnabled = req.TopPEnabled
	if req.TopP != nil {
		t := float64(*req.TopP)
		params.TopP = &t
	}
	params.TopKEnabled = req.TopKEnabled
	if req.TopK != nil {
		t := int(*req.TopK)
		params.TopK = &t
	}
	if req.MaxToolCalls != nil {
		t := int(*req.MaxToolCalls)
		params.MaxToolCalls = &t
	}
	return params
}

func toListAgentsParams(params gen.ListAgentsParams) queries.ListAgentsParams {
	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	return queries.ListAgentsParams{
		Limit:  limit,
		Offset: offset,
	}
}

func toPlaygroundSessionResponse(s *playground.Session) gen.ModelsPlaygroundSessionResponse {
	return gen.ModelsPlaygroundSessionResponse{
		Id:        s.ID,
		AgentId:   s.AgentID,
		Title:     s.Title,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func toPlaygroundSessionListResponse(result queries.ListPlaygroundSessionsResult) gen.ModelsPlaygroundSessionListResponse {
	sessions := make([]gen.ModelsPlaygroundSessionResponse, len(result.Sessions))
	for i, s := range result.Sessions {
		sessions[i] = toPlaygroundSessionResponse(s)
	}
	return gen.ModelsPlaygroundSessionListResponse{Sessions: sessions}
}

func toPlaygroundMessageResponse(m *playground.Message) gen.ModelsPlaygroundMessageResponse {
	return gen.ModelsPlaygroundMessageResponse{
		Id:        m.ID,
		SessionId: m.SessionID,
		Role:      gen.ModelsPlaygroundMessageResponseRole(m.Role),
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

func toPlaygroundMessageListResponse(result queries.ListPlaygroundMessagesResult) gen.ModelsPlaygroundMessageListResponse {
	messages := make([]gen.ModelsPlaygroundMessageResponse, len(result.Messages))
	for i, m := range result.Messages {
		messages[i] = toPlaygroundMessageResponse(m)
	}
	return gen.ModelsPlaygroundMessageListResponse{Messages: messages}
}

// escapeSSEData escapes newlines in SSE data fields.
// Each line in an SSE data field must be prefixed with "data: ".
func escapeSSEData(s string) string {
	return strings.ReplaceAll(s, "\n", "\ndata: ")
}

// HTTP tool type conversion helpers

func toHttpToolHeaderResponse(h httptool.HttpToolHeader) gen.ModelsHttpToolHeader {
	return gen.ModelsHttpToolHeader{
		Name:  h.Name,
		Value: h.Value,
	}
}

func toHttpToolResponse(t *httptool.HttpTool) gen.ModelsHttpToolResponse {
	headers := make([]gen.ModelsHttpToolHeader, len(t.Headers))
	for i, h := range t.Headers {
		headers[i] = toHttpToolHeaderResponse(h)
	}
	return gen.ModelsHttpToolResponse{
		Id:           t.ID,
		Name:         t.Name,
		Description:  t.Description,
		Method:       gen.ModelsHttpMethod(t.Method),
		Url:          t.URL,
		Headers:      headers,
		BodyTemplate: t.BodyTemplate,
		InputSchema:  t.InputSchema,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

func toHttpToolListResponse(result queries.ListHttpToolsResult, limit, offset int) gen.ModelsHttpToolListResponse {
	tools := make([]gen.ModelsHttpToolResponse, len(result.HttpTools))
	for i, t := range result.HttpTools {
		tools[i] = toHttpToolResponse(t)
	}
	return gen.ModelsHttpToolListResponse{
		HttpTools: tools,
		Total:     int32(result.Total),
		Limit:     int32(limit),
		Offset:    int32(offset),
	}
}

func toCreateHttpToolParams(req gen.ModelsCreateHttpToolRequest) commands.CreateHttpToolParams {
	params := commands.CreateHttpToolParams{
		Name:   req.Name,
		Method: string(req.Method),
		URL:    req.Url,
	}
	if req.Description != nil {
		params.Description = *req.Description
	}
	if req.Headers != nil {
		headers := make([]httptool.HttpToolHeader, len(*req.Headers))
		for i, h := range *req.Headers {
			headers[i] = httptool.HttpToolHeader{Name: h.Name, Value: h.Value}
		}
		params.Headers = headers
	}
	if req.BodyTemplate != nil {
		params.BodyTemplate = *req.BodyTemplate
	}
	if req.InputSchema != nil {
		params.InputSchema = *req.InputSchema
	}
	return params
}

func toUpdateHttpToolParams(id string, req gen.ModelsUpdateHttpToolRequest) commands.UpdateHttpToolParams {
	params := commands.UpdateHttpToolParams{
		ID:           id,
		Name:         req.Name,
		Description:  req.Description,
		BodyTemplate: req.BodyTemplate,
		InputSchema:  req.InputSchema,
		URL:          req.Url,
	}
	if req.Method != nil {
		m := string(*req.Method)
		params.Method = &m
	}
	if req.Headers != nil {
		headers := make([]httptool.HttpToolHeader, len(*req.Headers))
		for i, h := range *req.Headers {
			headers[i] = httptool.HttpToolHeader{Name: h.Name, Value: h.Value}
		}
		params.Headers = &headers
	}
	return params
}

func toListHttpToolsParams(params gen.ListHttpToolsParams) queries.ListHttpToolsParams {
	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	return queries.ListHttpToolsParams{
		Limit:  limit,
		Offset: offset,
	}
}

// Deploy endpoint type conversion helpers

func toDeployChatMessages(messages []gen.ModelsDeployChatMessage) []chatprovider.ChatMessage {
	out := make([]chatprovider.ChatMessage, 0, len(messages))
	for _, m := range messages {
		out = append(out, chatprovider.ChatMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}
	return out
}

func toOpenAIChatCompletion(id, model, content string, created int64) gen.ModelsDeployChatCompletionResponse {
	return gen.ModelsDeployChatCompletionResponse{
		Id:      id,
		Object:  "chat.completion",
		Created: created,
		Model:   model,
		Choices: []gen.ModelsDeployChatCompletionChoice{
			{
				Index: 0,
				Message: gen.ModelsDeployChatMessage{
					Role:    gen.ModelsDeployChatMessageRoleAssistant,
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: gen.ModelsDeployChatCompletionUsage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}
}

// SSE streaming chunk types (not in generated types since SSE isn't modeled in TypeSpec)

type deployChatCompletionChunk struct {
	ID      string                              `json:"id"`
	Object  string                              `json:"object"`
	Created int64                               `json:"created"`
	Model   string                              `json:"model"`
	Choices []deployChatCompletionChunkChoice    `json:"choices"`
}

type deployChatCompletionChunkChoice struct {
	Index        int             `json:"index"`
	Delta        deployChatDelta `json:"delta"`
	FinishReason *string         `json:"finish_reason"`
}

type deployChatDelta struct {
	Role    *string `json:"role,omitempty"`
	Content *string `json:"content,omitempty"`
}

func hasSystemMessage(messages []gen.ModelsDeployChatMessage) bool {
	for _, m := range messages {
		if m.Role == gen.ModelsDeployChatMessageRoleSystem {
			return true
		}
	}
	return false
}

func toDeploySessionResponse(s *playground.Session) gen.ModelsDeploySessionResponse {
	return gen.ModelsDeploySessionResponse{
		Id:        s.ID,
		AgentId:   s.AgentID,
		Title:     s.Title,
		CreatedAt: s.CreatedAt,
	}
}

// ---- MCP Server type conversions ----

func toMcpServerHeaderResponse(h mcpserver.McpServerHeader) gen.ModelsMcpServerHeader {
	return gen.ModelsMcpServerHeader{
		Name:  h.Name,
		Value: h.Value,
	}
}

func toMcpServerResponse(s *mcpserver.McpServer) gen.ModelsMcpServerResponse {
	headers := make([]gen.ModelsMcpServerHeader, len(s.Headers))
	for i, h := range s.Headers {
		headers[i] = toMcpServerHeaderResponse(h)
	}
	return gen.ModelsMcpServerResponse{
		Id:        s.ID,
		Name:      s.Name,
		Transport: gen.ModelsMcpTransport(s.Transport),
		Url:       s.URL,
		Headers:   headers,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func toMcpServerListResponse(result queries.ListMcpServersResult, limit, offset int) gen.ModelsMcpServerListResponse {
	servers := make([]gen.ModelsMcpServerResponse, len(result.McpServers))
	for i, s := range result.McpServers {
		servers[i] = toMcpServerResponse(s)
	}
	return gen.ModelsMcpServerListResponse{
		McpServers: servers,
		Total:      int32(result.Total),
		Limit:      int32(limit),
		Offset:     int32(offset),
	}
}

func toCreateMcpServerParams(req gen.ModelsCreateMcpServerRequest) commands.CreateMcpServerParams {
	params := commands.CreateMcpServerParams{
		Name:      req.Name,
		Transport: string(req.Transport),
		URL:       req.Url,
	}
	if req.Headers != nil {
		headers := make([]mcpserver.McpServerHeader, len(*req.Headers))
		for i, h := range *req.Headers {
			headers[i] = mcpserver.McpServerHeader{Name: h.Name, Value: h.Value}
		}
		params.Headers = headers
	}
	return params
}

func toUpdateMcpServerParams(id string, req gen.ModelsUpdateMcpServerRequest) commands.UpdateMcpServerParams {
	params := commands.UpdateMcpServerParams{
		ID:  id,
		Name: req.Name,
		URL:  req.Url,
	}
	if req.Transport != nil {
		t := string(*req.Transport)
		params.Transport = &t
	}
	if req.Headers != nil {
		headers := make([]mcpserver.McpServerHeader, len(*req.Headers))
		for i, h := range *req.Headers {
			headers[i] = mcpserver.McpServerHeader{Name: h.Name, Value: h.Value}
		}
		params.Headers = &headers
	}
	return params
}

func toListMcpServersParams(params gen.ListMcpServersParams) queries.ListMcpServersParams {
	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	return queries.ListMcpServersParams{
		Limit:  limit,
		Offset: offset,
	}
}

// --- Agent Tool Conversions ---

func toAgentToolConfigResponse(config agenttoolrepo.AgentToolConfig) gen.ModelsAgentToolConfigResponse {
	httpToolIDs := config.HttpToolIDs
	if httpToolIDs == nil {
		httpToolIDs = []string{}
	}

	builtInToolIDs := config.BuiltInToolIDs
	if builtInToolIDs == nil {
		builtInToolIDs = []string{}
	}

	mcpServers := make([]gen.ModelsAgentMcpServerToolConfig, 0, len(config.McpServers))
	for _, mc := range config.McpServers {
		toolNames := mc.ToolNames
		if toolNames == nil {
			toolNames = []string{}
		}
		mcpServers = append(mcpServers, gen.ModelsAgentMcpServerToolConfig{
			McpServerId: mc.McpServerID,
			SelectAll:   mc.SelectAll,
			ToolNames:   toolNames,
		})
	}

	return gen.ModelsAgentToolConfigResponse{
		HttpToolIds:    httpToolIDs,
		McpServers:     mcpServers,
		BuiltInToolIds: builtInToolIDs,
	}
}

func toUpdateAgentToolsParams(agentID string, req gen.ModelsUpdateAgentToolConfigRequest) commands.UpdateAgentToolsParams {
	mcpServers := make([]agenttoolrepo.McpServerToolConfig, 0, len(req.McpServers))
	for _, mc := range req.McpServers {
		toolNames := mc.ToolNames
		if toolNames == nil {
			toolNames = []string{}
		}
		mcpServers = append(mcpServers, agenttoolrepo.McpServerToolConfig{
			McpServerID: mc.McpServerId,
			SelectAll:   mc.SelectAll,
			ToolNames:   toolNames,
		})
	}

	builtInToolIDs := req.BuiltInToolIds
	if builtInToolIDs == nil {
		builtInToolIDs = []string{}
	}

	return commands.UpdateAgentToolsParams{
		AgentID:        agentID,
		HttpToolIDs:    req.HttpToolIds,
		McpServers:     mcpServers,
		BuiltInToolIDs: builtInToolIDs,
	}
}

func toToolOverride(httpToolIDs *[]string, mcpServers *[]gen.ModelsPlaygroundMcpServerToolOverride, builtInToolIDs *[]string) *queries.ToolOverride {
	override := &queries.ToolOverride{}

	if httpToolIDs != nil {
		override.HttpToolIDs = *httpToolIDs
	}

	if builtInToolIDs != nil {
		override.BuiltInToolIDs = *builtInToolIDs
	}

	if mcpServers != nil {
		for _, mc := range *mcpServers {
			toolNames := mc.ToolNames
			if toolNames == nil {
				toolNames = []string{}
			}
			override.McpServers = append(override.McpServers, agenttoolrepo.McpServerToolConfig{
				McpServerID: mc.McpServerId,
				SelectAll:   mc.SelectAll,
				ToolNames:   toolNames,
			})
		}
	}

	return override
}

func toBuiltInToolListResponse(result queries.ListBuiltInToolsResult) gen.ModelsBuiltInToolListResponse {
	tools := make([]gen.ModelsBuiltInToolResponse, len(result.BuiltInTools))
	for i, t := range result.BuiltInTools {
		tools[i] = toBuiltInToolResponse(t)
	}
	return gen.ModelsBuiltInToolListResponse{
		BuiltInTools: tools,
	}
}

func toBuiltInToolResponse(t builtintool.BuiltInTool) gen.ModelsBuiltInToolResponse {
	return gen.ModelsBuiltInToolResponse{
		Id:          t.ID,
		Name:        t.Name,
		Description: t.Description,
	}
}
