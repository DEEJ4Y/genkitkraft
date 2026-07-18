package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
	mcpdiscovery "github.com/DEEJ4Y/genkitkraft/internal/ports/mcp_discovery"
)

const sessionCookieName = "session_token"
const sessionMaxAge = 86400 // 24 hours

// Compile-time check that Handler implements the generated ServerInterface.
var _ gen.ServerInterface = (*Handler)(nil)

// Handler implements gen.ServerInterface, delegating to the application layer.
type Handler struct {
	authApp        *app.AuthApp
	providerApp    *app.ProviderApp
	promptApp      *app.PromptApp
	agentApp       *app.AgentApp
	playgroundApp  *app.PlaygroundApp
	httpToolApp    *app.HttpToolApp
	mcpServerApp   *app.McpServerApp
	agentToolApp   *app.AgentToolApp
	builtInToolApp *app.BuiltInToolApp
	chatProvider   chatprovider.ChatProvider
	mcpDiscovery   mcpdiscovery.McpDiscovery
}

func NewHandler(authApp *app.AuthApp, providerApp *app.ProviderApp, promptApp *app.PromptApp, agentApp *app.AgentApp, playgroundApp *app.PlaygroundApp, httpToolApp *app.HttpToolApp, mcpServerApp *app.McpServerApp, agentToolApp *app.AgentToolApp, builtInToolApp *app.BuiltInToolApp, chatProvider chatprovider.ChatProvider, mcpDiscovery mcpdiscovery.McpDiscovery) *Handler {
	return &Handler{authApp: authApp, providerApp: providerApp, promptApp: promptApp, agentApp: agentApp, playgroundApp: playgroundApp, httpToolApp: httpToolApp, mcpServerApp: mcpServerApp, agentToolApp: agentToolApp, builtInToolApp: builtInToolApp, chatProvider: chatProvider, mcpDiscovery: mcpDiscovery}
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

func (h *Handler) ListHttpTools(w http.ResponseWriter, r *http.Request, params gen.ListHttpToolsParams) {
	qParams := toListHttpToolsParams(params)
	result, err := h.httpToolApp.Queries.ListHttpTools.Execute(r.Context(), qParams)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toHttpToolListResponse(result, qParams.Limit, qParams.Offset))
}

func (h *Handler) CreateHttpTool(w http.ResponseWriter, r *http.Request) {
	var req gen.ModelsCreateHttpToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := toCreateHttpToolParams(req)
	result, err := h.httpToolApp.Commands.CreateHttpTool.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toHttpToolResponse(result.HttpTool))
}

func (h *Handler) GetHttpTool(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.httpToolApp.Queries.GetHttpTool.Execute(r.Context(), queries.GetHttpToolParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toHttpToolResponse(result.HttpTool))
}

func (h *Handler) UpdateHttpTool(w http.ResponseWriter, r *http.Request, id string) {
	var req gen.ModelsUpdateHttpToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := toUpdateHttpToolParams(id, req)
	result, err := h.httpToolApp.Commands.UpdateHttpTool.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toHttpToolResponse(result.HttpTool))
}

func (h *Handler) DeleteHttpTool(w http.ResponseWriter, r *http.Request, id string) {
	err := h.httpToolApp.Commands.DeleteHttpTool.Execute(r.Context(), commands.DeleteHttpToolParams{ID: id})
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
	err := h.playgroundApp.Commands.DeleteSession.Execute(r.Context(), commands.DeletePlaygroundSessionParams{ID: sessionId, AgentID: agentId})
	if err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListPlaygroundMessages(w http.ResponseWriter, r *http.Request, agentId string, sessionId string) {
	result, err := h.playgroundApp.Queries.ListMessages.Execute(r.Context(), queries.ListPlaygroundMessagesParams{SessionID: sessionId, AgentID: agentId})
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
	if req.TemperatureEnabled != nil {
		configParams.TemperatureEnabled = req.TemperatureEnabled
	}
	if req.TopP != nil {
		t := float64(*req.TopP)
		configParams.TopP = &t
	}
	if req.TopPEnabled != nil {
		configParams.TopPEnabled = req.TopPEnabled
	}
	if req.TopK != nil {
		t := int(*req.TopK)
		configParams.TopK = &t
	}
	if req.TopKEnabled != nil {
		configParams.TopKEnabled = req.TopKEnabled
	}
	if req.MaxToolCalls != nil {
		t := int(*req.MaxToolCalls)
		configParams.MaxToolCalls = &t
	}

	// Tool overrides
	configParams.IncludeTools = true
	if req.HttpToolIds != nil || req.McpServers != nil || req.BuiltInToolIds != nil {
		configParams.ToolOverride = toToolOverride(req.HttpToolIds, req.McpServers, req.BuiltInToolIds)
	}

	configResult, err := h.playgroundApp.Queries.ResolveConfig.Execute(r.Context(), configParams)
	if err != nil {
		writeAppError(w, err)
		return
	}

	// Load conversation history
	messagesResult, err := h.playgroundApp.Queries.ListMessages.Execute(r.Context(), queries.ListPlaygroundMessagesParams{SessionID: req.SessionId, AgentID: agentId})
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

	// Non-streaming path: return a single JSON response
	if req.Stream != nil && !*req.Stream {
		content, err := h.chatProvider.Chat(r.Context(), chatReq)
		if err != nil {
			writeAppError(w, err)
			return
		}

		if content != "" {
			result, saveErr := h.playgroundApp.Commands.SaveMessage.Execute(r.Context(), commands.SavePlaygroundMessageParams{
				SessionID: req.SessionId,
				Role:      "assistant",
				Content:   content,
			})
			if saveErr == nil {
				writeJSON(w, http.StatusOK, toPlaygroundMessageResponse(result.Message))
				return
			}
		}

		writeJSON(w, http.StatusOK, gen.ModelsPlaygroundMessageResponse{
			SessionId: req.SessionId,
			Role:      gen.ModelsPlaygroundMessageResponseRoleAssistant,
			Content:   content,
		})
		return
	}

	// Streaming path: SSE response. Generation runs detached from this
	// request (see StartPlaygroundStreamCommand) so a client disconnect
	// doesn't abort it; this handler just tails the persisted chunks.
	startResult, err := h.playgroundApp.Commands.StartStream.Execute(r.Context(), commands.StartPlaygroundStreamParams{
		SessionID:   req.SessionId,
		ChatRequest: chatReq,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	h.streamMessage(w, r, req.SessionId, startResult.MessageID, 0)
}

// PlaygroundChatStream reconnects to the assistant reply currently (or most
// recently) streaming for a session. A Last-Event-ID header resumes from the
// next token instead of replaying the whole reply.
func (h *Handler) PlaygroundChatStream(w http.ResponseWriter, r *http.Request, agentId string, sessionId string) {
	if _, err := h.playgroundApp.Queries.GetSession.Execute(r.Context(), queries.GetPlaygroundSessionParams{
		SessionID: sessionId,
		AgentID:   agentId,
	}); err != nil {
		writeAppError(w, err)
		return
	}

	h.streamMessage(w, r, sessionId, "", parseLastEventID(r))
}

// CancelPlaygroundStream stops the assistant reply currently streaming for a
// session, if any. Best-effort: a no-op if generation already finished or is
// running on a different instance.
func (h *Handler) CancelPlaygroundStream(w http.ResponseWriter, r *http.Request, agentId string, sessionId string) {
	if _, err := h.playgroundApp.Queries.GetSession.Execute(r.Context(), queries.GetPlaygroundSessionParams{
		SessionID: sessionId,
		AgentID:   agentId,
	}); err != nil {
		writeAppError(w, err)
		return
	}

	if err := h.playgroundApp.Commands.CancelStream.Execute(r.Context(), commands.CancelPlaygroundStreamParams{
		SessionID: sessionId,
	}); err != nil {
		writeAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeployChatCompletions(w http.ResponseWriter, r *http.Request, agentId string) {
	var req gen.ModelsDeployChatCompletionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	if len(req.Messages) == 0 {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "messages array must not be empty"))
		return
	}

	if hasSystemMessage(req.Messages) {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "system messages are not allowed; the agent's system prompt is applied server-side"))
		return
	}

	// Resolve agent config (provider, model, system prompt, tools)
	configResult, err := h.playgroundApp.Queries.ResolveConfig.Execute(r.Context(), queries.ResolvePlaygroundConfigParams{
		AgentID:      agentId,
		IncludeTools: true,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	chatReq := configResult.ChatRequest
	chatReq.Messages = toDeployChatMessages(req.Messages)

	completionID := "chatcmpl-" + uuid.New().String()
	created := time.Now().Unix()
	model := chatReq.ModelID

	// Non-streaming response
	if req.Stream == nil || !*req.Stream {
		content, err := h.chatProvider.Chat(r.Context(), chatReq)
		if err != nil {
			writeAppError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, toOpenAIChatCompletion(completionID, model, content, created))
		return
	}

	// Streaming response: do all validation above this point
	tokenCh, errCh := h.chatProvider.ChatStream(r.Context(), chatReq)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeAppError(w, errors.NewAppError(errors.Internal, "streaming not supported"))
		return
	}

	// Initial chunk with role
	roleChunk := deployChatCompletionChunk{
		ID:      completionID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []deployChatCompletionChunkChoice{
			{
				Index: 0,
				Delta: deployChatDelta{Role: strPtr("assistant")},
			},
		},
	}
	roleData, _ := json.Marshal(roleChunk)
	fmt.Fprintf(w, "data: %s\n\n", roleData)
	flusher.Flush()

	// Content chunks
	for token := range tokenCh {
		chunk := deployChatCompletionChunk{
			ID:      completionID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   model,
			Choices: []deployChatCompletionChunkChoice{
				{
					Index: 0,
					Delta: deployChatDelta{Content: &token},
				},
			},
		}
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	if streamErr := <-errCh; streamErr != nil {
		errData, _ := json.Marshal(map[string]interface{}{
			"error": map[string]string{
				"message": streamErr.Error(),
				"type":    "server_error",
			},
		})
		fmt.Fprintf(w, "data: %s\n\n", errData)
		flusher.Flush()
		return
	}

	// Final chunk with finish_reason
	stopReason := "stop"
	finishChunk := deployChatCompletionChunk{
		ID:      completionID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []deployChatCompletionChunkChoice{
			{
				Index:        0,
				Delta:        deployChatDelta{},
				FinishReason: &stopReason,
			},
		},
	}
	finishData, _ := json.Marshal(finishChunk)
	fmt.Fprintf(w, "data: %s\n\n", finishData)
	flusher.Flush()

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (h *Handler) CreateDeploySession(w http.ResponseWriter, r *http.Request, agentId string) {
	var req gen.ModelsCreateDeploySessionRequest
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

	writeJSON(w, http.StatusCreated, toDeploySessionResponse(result.Session))
}

func (h *Handler) GetDeploySession(w http.ResponseWriter, r *http.Request, agentId string, sessionId string) {
	result, err := h.playgroundApp.Queries.GetSession.Execute(r.Context(), queries.GetPlaygroundSessionParams{
		SessionID: sessionId,
		AgentID:   agentId,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toDeploySessionResponse(result.Session))
}

func (h *Handler) DeleteDeploySession(w http.ResponseWriter, r *http.Request, agentId string, sessionId string) {
	err := h.playgroundApp.Commands.DeleteSession.Execute(r.Context(), commands.DeletePlaygroundSessionParams{
		ID:      sessionId,
		AgentID: agentId,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeploySessionChatCompletions(w http.ResponseWriter, r *http.Request, agentId string, sessionId string) {
	var req gen.ModelsDeployChatCompletionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	if len(req.Messages) == 0 {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "messages array must not be empty"))
		return
	}

	if hasSystemMessage(req.Messages) {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "system messages are not allowed; the agent's system prompt is applied server-side"))
		return
	}

	lastMsg := req.Messages[len(req.Messages)-1]
	if lastMsg.Role != gen.ModelsDeployChatMessageRoleUser {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "the last message must have role 'user'"))
		return
	}

	// Validate session belongs to agent before saving
	_, err := h.playgroundApp.Queries.GetSession.Execute(r.Context(), queries.GetPlaygroundSessionParams{
		SessionID: sessionId,
		AgentID:   agentId,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	// Save the user message
	_, err = h.playgroundApp.Commands.SaveMessage.Execute(r.Context(), commands.SavePlaygroundMessageParams{
		SessionID: sessionId,
		Role:      "user",
		Content:   lastMsg.Content,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	// Load full conversation history (includes the just-saved user message)
	messagesResult, err := h.playgroundApp.Queries.ListMessages.Execute(r.Context(), queries.ListPlaygroundMessagesParams{
		SessionID: sessionId,
		AgentID:   agentId,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	// Resolve agent config
	configResult, err := h.playgroundApp.Queries.ResolveConfig.Execute(r.Context(), queries.ResolvePlaygroundConfigParams{
		AgentID: agentId,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	// Build chat request with full history
	chatMessages := make([]chatprovider.ChatMessage, 0, len(messagesResult.Messages))
	for _, m := range messagesResult.Messages {
		chatMessages = append(chatMessages, chatprovider.ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	chatReq := configResult.ChatRequest
	chatReq.Messages = chatMessages

	completionID := "chatcmpl-" + uuid.New().String()
	created := time.Now().Unix()
	model := chatReq.ModelID

	// Non-streaming response
	if req.Stream == nil || !*req.Stream {
		content, err := h.chatProvider.Chat(r.Context(), chatReq)
		if err != nil {
			writeAppError(w, err)
			return
		}

		// Save assistant reply — fail the request if persistence fails
		_, err = h.playgroundApp.Commands.SaveMessage.Execute(r.Context(), commands.SavePlaygroundMessageParams{
			SessionID: sessionId,
			Role:      "assistant",
			Content:   content,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, toOpenAIChatCompletion(completionID, model, content, created))
		return
	}

	// Streaming response. Generation runs detached from this request (see
	// StartPlaygroundStreamCommand) so a client disconnect doesn't abort it;
	// this handler just tails the persisted chunks.
	startResult, err := h.playgroundApp.Commands.StartStream.Execute(r.Context(), commands.StartPlaygroundStreamParams{
		SessionID:   sessionId,
		ChatRequest: chatReq,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeAppError(w, errors.NewAppError(errors.Internal, "streaming not supported"))
		return
	}

	// Initial chunk with role
	roleChunk := deployChatCompletionChunk{
		ID:      completionID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []deployChatCompletionChunkChoice{
			{
				Index: 0,
				Delta: deployChatDelta{Role: strPtr("assistant")},
			},
		},
	}
	roleData, _ := json.Marshal(roleChunk)
	fmt.Fprintf(w, "data: %s\n\n", roleData)
	flusher.Flush()

	h.streamDeployCompletion(w, r, sessionId, startResult.MessageID, completionID, model, created, 0)
}

// DeploySessionChatCompletionsStream reconnects to the assistant reply
// currently (or most recently) streaming for a session. A Last-Event-ID
// header resumes from the next token instead of replaying the whole reply.
func (h *Handler) DeploySessionChatCompletionsStream(w http.ResponseWriter, r *http.Request, agentId string, sessionId string) {
	if _, err := h.playgroundApp.Queries.GetSession.Execute(r.Context(), queries.GetPlaygroundSessionParams{
		SessionID: sessionId,
		AgentID:   agentId,
	}); err != nil {
		writeAppError(w, err)
		return
	}

	completionID := "chatcmpl-" + uuid.New().String()
	created := time.Now().Unix()

	configResult, err := h.playgroundApp.Queries.ResolveConfig.Execute(r.Context(), queries.ResolvePlaygroundConfigParams{
		AgentID: agentId,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	h.streamDeployCompletion(w, r, sessionId, "", completionID, configResult.ChatRequest.ModelID, created, parseLastEventID(r))
}

// CancelDeploySessionChatCompletions stops the assistant reply currently
// streaming for a session, if any. Best-effort: a no-op if generation already
// finished or is running on a different instance.
func (h *Handler) CancelDeploySessionChatCompletions(w http.ResponseWriter, r *http.Request, agentId string, sessionId string) {
	if _, err := h.playgroundApp.Queries.GetSession.Execute(r.Context(), queries.GetPlaygroundSessionParams{
		SessionID: sessionId,
		AgentID:   agentId,
	}); err != nil {
		writeAppError(w, err)
		return
	}

	if err := h.playgroundApp.Commands.CancelStream.Execute(r.Context(), commands.CancelPlaygroundStreamParams{
		SessionID: sessionId,
	}); err != nil {
		writeAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func strPtr(s string) *string {
	return &s
}

// ---- MCP Server handlers ----

func (h *Handler) ListMcpServers(w http.ResponseWriter, r *http.Request, params gen.ListMcpServersParams) {
	qParams := toListMcpServersParams(params)
	result, err := h.mcpServerApp.Queries.ListMcpServers.Execute(r.Context(), qParams)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toMcpServerListResponse(result, qParams.Limit, qParams.Offset))
}

func (h *Handler) CreateMcpServer(w http.ResponseWriter, r *http.Request) {
	var req gen.ModelsCreateMcpServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := toCreateMcpServerParams(req)
	result, err := h.mcpServerApp.Commands.CreateMcpServer.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toMcpServerResponse(result.McpServer))
}

func (h *Handler) GetMcpServer(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.mcpServerApp.Queries.GetMcpServer.Execute(r.Context(), queries.GetMcpServerParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toMcpServerResponse(result.McpServer))
}

func (h *Handler) UpdateMcpServer(w http.ResponseWriter, r *http.Request, id string) {
	var req gen.ModelsUpdateMcpServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	params := toUpdateMcpServerParams(id, req)
	result, err := h.mcpServerApp.Commands.UpdateMcpServer.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toMcpServerResponse(result.McpServer))
}

func (h *Handler) DeleteMcpServer(w http.ResponseWriter, r *http.Request, id string) {
	err := h.mcpServerApp.Commands.DeleteMcpServer.Execute(r.Context(), commands.DeleteMcpServerParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListMcpServerTools(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.mcpServerApp.Queries.GetMcpServer.Execute(r.Context(), queries.GetMcpServerParams{ID: id})
	if err != nil {
		writeAppError(w, err)
		return
	}

	server := result.McpServer
	headerMap := make(map[string]string, len(server.Headers))
	for _, hdr := range server.Headers {
		headerMap[hdr.Name] = hdr.Value
	}

	tools, err := h.mcpDiscovery.ListTools(r.Context(), server.Name, server.Transport, server.URL, headerMap)
	if err != nil {
		writeJSON(w, http.StatusOK, gen.ModelsMcpServerToolListResponse{
			Tools: []gen.ModelsMcpServerToolResponse{},
		})
		return
	}

	resp := make([]gen.ModelsMcpServerToolResponse, len(tools))
	for i, t := range tools {
		resp[i] = gen.ModelsMcpServerToolResponse{
			Name:        t.Name,
			Description: t.Description,
		}
	}
	writeJSON(w, http.StatusOK, gen.ModelsMcpServerToolListResponse{Tools: resp})
}

// --- Agent Tools ---

func (h *Handler) GetAgentTools(w http.ResponseWriter, r *http.Request, agentId string) {
	result, err := h.agentToolApp.Queries.GetTools.Execute(r.Context(), queries.GetAgentToolsParams{
		AgentID: agentId,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAgentToolConfigResponse(result.Config))
}

func (h *Handler) UpdateAgentTools(w http.ResponseWriter, r *http.Request, agentId string) {
	var req gen.ModelsUpdateAgentToolConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	result, err := h.agentToolApp.Commands.UpdateTools.Execute(r.Context(), toUpdateAgentToolsParams(agentId, req))
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAgentToolConfigResponse(result.Config))
}

func (h *Handler) ListBuiltInTools(w http.ResponseWriter, r *http.Request) {
	result, err := h.builtInToolApp.Queries.ListBuiltInTools.Execute(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBuiltInToolListResponse(result))
}