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
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
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
		Id:           a.ID,
		Name:         a.Name,
		ProviderId:   a.ProviderID,
		ProviderName: a.ProviderName,
		ProviderType: gen.ModelsProviderType(a.ProviderType),
		ModelId:      a.ModelID,
		Temperature:  float32(a.Temperature),
		TopP:         float32(a.TopP),
		TopK:         int32(a.TopK),
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
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
	if req.Temperature != nil {
		t := float64(*req.Temperature)
		params.Temperature = &t
	}
	if req.TopP != nil {
		t := float64(*req.TopP)
		params.TopP = &t
	}
	if req.TopK != nil {
		t := int(*req.TopK)
		params.TopK = &t
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
	if req.Temperature != nil {
		t := float64(*req.Temperature)
		params.Temperature = &t
	}
	if req.TopP != nil {
		t := float64(*req.TopP)
		params.TopP = &t
	}
	if req.TopK != nil {
		t := int(*req.TopK)
		params.TopK = &t
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
