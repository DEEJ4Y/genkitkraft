package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
)

const sessionCookieName = "session_token"
const sessionMaxAge = 86400 // 24 hours

// Compile-time check that Handler implements the generated ServerInterface.
var _ gen.ServerInterface = (*Handler)(nil)

// Handler implements gen.ServerInterface, delegating to the application layer.
type Handler struct {
	authApp       *app.AuthApp
	providerApp   *app.ProviderApp
	promptApp     *app.PromptApp
	agentApp      *app.AgentApp
	playgroundApp *app.PlaygroundApp
	chatProvider  chatprovider.ChatProvider
}

func NewHandler(authApp *app.AuthApp, providerApp *app.ProviderApp, promptApp *app.PromptApp, agentApp *app.AgentApp, playgroundApp *app.PlaygroundApp, chatProvider chatprovider.ChatProvider) *Handler {
	return &Handler{authApp: authApp, providerApp: providerApp, promptApp: promptApp, agentApp: agentApp, playgroundApp: playgroundApp, chatProvider: chatProvider}
}

func (h *Handler) GetAuthStatus(w http.ResponseWriter, r *http.Request) {
	result, err := h.authApp.Queries.GetAuthStatus.Execute(r.Context(), queries.GetAuthStatusParams{})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAuthStatusResponse(result))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req gen.ModelsLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	if req.Username == "" || req.Password == "" {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "username and password are required"))
		return
	}

	params := toLoginParams(req, extractIP(r))
	result, err := h.authApp.Commands.Login.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    result.Token,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   isSecure(r),
	})

	writeJSON(w, http.StatusOK, toLoginResponse(result))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		_ = h.authApp.Commands.Logout.Execute(r.Context(), commands.LogoutParams{Token: cookie.Value})
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   isSecure(r),
	})

	writeJSON(w, http.StatusOK, gen.ModelsLogoutResponse{Ok: true})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeAppError(w, errors.NewAppError(errors.Unauthorized, "unauthorized"))
		return
	}

	result, err := h.authApp.Queries.GetMe.Execute(r.Context(), queries.GetMeParams{Token: cookie.Value})
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toMeResponse(result))
}

func (h *Handler) HealthLiveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, gen.ModelsHealthCheckResponse{
		Status:  gen.Up,
		Message: "Server is running",
	})
}

func (h *Handler) HealthReadiness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, gen.ModelsHealthCheckResponse{
		Status:  gen.Up,
		Message: "Server is ready",
	})
}

func (h *Handler) ListProviderTypes(w http.ResponseWriter, r *http.Request) {
	result, err := h.providerApp.Queries.ListProviderTypes.Execute(r.Context(), queries.ListProviderTypesParams{})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProviderTypeListResponse(result))
}

func (h *Handler) ListProviders(w http.ResponseWriter, r *http.Request) {
	result, err := h.providerApp.Queries.ListProviders.Execute(r.Context(), queries.ListProvidersParams{})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProviderListResponse(result))
}

func (h *Handler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var req gen.ModelsCreateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := toCreateProviderParams(req)
	result, err := h.providerApp.Commands.CreateProvider.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toProviderResponse(result.Provider))
}

func (h *Handler) GetProvider(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.providerApp.Queries.GetProvider.Execute(r.Context(), queries.GetProviderParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProviderResponse(result.Provider))
}

func (h *Handler) UpdateProvider(w http.ResponseWriter, r *http.Request, id string) {
	var req gen.ModelsUpdateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := toUpdateProviderParams(id, req)
	result, err := h.providerApp.Commands.UpdateProvider.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toProviderResponse(result.Provider))
}

func (h *Handler) DeleteProvider(w http.ResponseWriter, r *http.Request, id string) {
	err := h.providerApp.Commands.DeleteProvider.Execute(r.Context(), commands.DeleteProviderParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) TestProvider(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.providerApp.Commands.TestProvider.Execute(r.Context(), commands.TestProviderParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, gen.ModelsTestProviderResponse{
		Success: result.Success,
		Message: result.Message,
	})
}

func (h *Handler) ListPrompts(w http.ResponseWriter, r *http.Request, params gen.ListPromptsParams) {
	qParams := toListPromptsParams(params)
	result, err := h.promptApp.Queries.ListPrompts.Execute(r.Context(), qParams)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPromptListResponse(result, qParams.Limit, qParams.Offset))
}

func (h *Handler) CreatePrompt(w http.ResponseWriter, r *http.Request) {
	var req gen.ModelsCreatePromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := toCreatePromptParams(req)
	result, err := h.promptApp.Commands.CreatePrompt.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPromptResponse(result.Prompt))
}

func (h *Handler) GetPrompt(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.promptApp.Queries.GetPrompt.Execute(r.Context(), queries.GetPromptParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPromptResponse(result.Prompt))
}

func (h *Handler) UpdatePrompt(w http.ResponseWriter, r *http.Request, id string) {
	var req gen.ModelsUpdatePromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := toUpdatePromptParams(id, req)
	result, err := h.promptApp.Commands.UpdatePrompt.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPromptResponse(result.Prompt))
}

func (h *Handler) DeletePrompt(w http.ResponseWriter, r *http.Request, id string) {
	err := h.promptApp.Commands.DeletePrompt.Execute(r.Context(), commands.DeletePromptParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListAgents(w http.ResponseWriter, r *http.Request, params gen.ListAgentsParams) {
	qParams := toListAgentsParams(params)
	result, err := h.agentApp.Queries.ListAgents.Execute(r.Context(), qParams)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAgentListResponse(result, qParams.Limit, qParams.Offset))
}

func (h *Handler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req gen.ModelsCreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := toCreateAgentParams(req)
	result, err := h.agentApp.Commands.CreateAgent.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toAgentResponse(result.Agent))
}

func (h *Handler) GetAgent(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.agentApp.Queries.GetAgent.Execute(r.Context(), queries.GetAgentParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAgentResponse(result.Agent))
}

func (h *Handler) UpdateAgent(w http.ResponseWriter, r *http.Request, id string) {
	var req gen.ModelsUpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := toUpdateAgentParams(id, req)
	result, err := h.agentApp.Commands.UpdateAgent.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAgentResponse(result.Agent))
}

func (h *Handler) DeleteAgent(w http.ResponseWriter, r *http.Request, id string) {
	err := h.agentApp.Commands.DeleteAgent.Execute(r.Context(), commands.DeleteAgentParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListPlaygroundSessions(w http.ResponseWriter, r *http.Request, agentId string) {
	result, err := h.playgroundApp.Queries.ListSessions.Execute(r.Context(), queries.ListPlaygroundSessionsParams{AgentID: agentId})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPlaygroundSessionListResponse(result))
}

func (h *Handler) CreatePlaygroundSession(w http.ResponseWriter, r *http.Request, agentId string) {
	var req gen.ModelsCreatePlaygroundSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := commands.CreatePlaygroundSessionParams{
		AgentID: agentId,
	}
	if req.Title != nil {
		params.Title = *req.Title
	}

	result, err := h.playgroundApp.Commands.CreateSession.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPlaygroundSessionResponse(result.Session))
}

func (h *Handler) DeletePlaygroundSession(w http.ResponseWriter, r *http.Request, agentId string, sessionId string) {
	err := h.playgroundApp.Commands.DeleteSession.Execute(r.Context(), commands.DeletePlaygroundSessionParams{ID: sessionId})
	if err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListPlaygroundMessages(w http.ResponseWriter, r *http.Request, agentId string, sessionId string) {
	result, err := h.playgroundApp.Queries.ListMessages.Execute(r.Context(), queries.ListPlaygroundMessagesParams{SessionID: sessionId})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPlaygroundMessageListResponse(result))
}

func (h *Handler) PlaygroundChat(w http.ResponseWriter, r *http.Request, agentId string) {
	var req gen.ModelsPlaygroundChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	if req.SessionId == "" {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "session ID is required"))
		return
	}
	if req.Content == "" {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "content is required"))
		return
	}

	// Save user message
	_, err := h.playgroundApp.Commands.SaveMessage.Execute(r.Context(), commands.SavePlaygroundMessageParams{
		SessionID: req.SessionId,
		Role:      "user",
		Content:   req.Content,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	// Resolve config (agent defaults + overrides)
	configParams := queries.ResolvePlaygroundConfigParams{
		AgentID: agentId,
	}
	if req.ProviderId != nil {
		configParams.ProviderID = *req.ProviderId
	}
	if req.ModelId != nil {
		configParams.ModelID = *req.ModelId
	}
	if req.SystemPromptId != nil {
		configParams.SystemPromptID = req.SystemPromptId
	}
	if req.Temperature != nil {
		t := float64(*req.Temperature)
		configParams.Temperature = &t
	}
	if req.TopP != nil {
		t := float64(*req.TopP)
		configParams.TopP = &t
	}
	if req.TopK != nil {
		t := int(*req.TopK)
		configParams.TopK = &t
	}

	configResult, err := h.playgroundApp.Queries.ResolveConfig.Execute(r.Context(), configParams)
	if err != nil {
		writeAppError(w, err)
		return
	}

	// Load conversation history
	messagesResult, err := h.playgroundApp.Queries.ListMessages.Execute(r.Context(), queries.ListPlaygroundMessagesParams{SessionID: req.SessionId})
	if err != nil {
		writeAppError(w, err)
		return
	}

	// Build chat messages from history
	chatMessages := make([]chatprovider.ChatMessage, 0, len(messagesResult.Messages))
	for _, m := range messagesResult.Messages {
		chatMessages = append(chatMessages, chatprovider.ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	chatReq := configResult.ChatRequest
	chatReq.Messages = chatMessages

	// Stream response via SSE
	tokenCh, errCh := h.chatProvider.ChatStream(r.Context(), chatReq)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeAppError(w, errors.NewAppError(errors.Internal, "streaming not supported"))
		return
	}

	var fullResponse strings.Builder

	for token := range tokenCh {
		fullResponse.WriteString(token)
		fmt.Fprintf(w, "data: %s\n\n", escapeSSEData(token))
		flusher.Flush()
	}

	// Check for streaming errors
	if streamErr := <-errCh; streamErr != nil {
		fmt.Fprintf(w, "data: [ERROR] %s\n\n", streamErr.Error())
		flusher.Flush()
		return
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	// Save assistant message
	if fullResponse.Len() > 0 {
		_, _ = h.playgroundApp.Commands.SaveMessage.Execute(r.Context(), commands.SavePlaygroundMessageParams{
			SessionID: req.SessionId,
			Role:      "assistant",
			Content:   fullResponse.String(),
		})
	}
}
