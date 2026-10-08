package mcphandler

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	agenttoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_tool_repo"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
)

// --- Input/Output types ---

type ListPlaygroundSessionsInput struct {
	AgentID string `json:"agent_id" jsonschema:"agent ID to list sessions for (required)"`
}

type PlaygroundSessionOutput struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListPlaygroundSessionsOutput struct {
	Sessions []PlaygroundSessionOutput `json:"sessions"`
}

type CreatePlaygroundSessionInput struct {
	AgentID string `json:"agent_id" jsonschema:"agent ID (required)"`
	Title   string `json:"title,omitempty" jsonschema:"session title (optional; auto-generated from the first message if omitted)"`
}

type DeletePlaygroundSessionInput struct {
	ID      string `json:"id" jsonschema:"session ID to delete (required)"`
	AgentID string `json:"agent_id" jsonschema:"agent ID (required)"`
}

type ListPlaygroundMessagesInput struct {
	SessionID string `json:"session_id" jsonschema:"session ID (required)"`
	AgentID   string `json:"agent_id" jsonschema:"agent ID (required)"`
}

type PlaygroundMessageOutput struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ListPlaygroundMessagesOutput struct {
	Messages []PlaygroundMessageOutput `json:"messages"`
}

type DeletePlaygroundSessionOutput struct {
	Status string `json:"status"`
}

type PlaygroundChatInput struct {
	AgentID   string `json:"agent_id" jsonschema:"agent ID (required)"`
	SessionID string `json:"session_id" jsonschema:"ID of an existing session for this agent (required); get one from playground_sessions_create or playground_sessions_list"`
	Content   string `json:"content" jsonschema:"user message content (required)"`

	// Optional per-request overrides. Omitted fields use the agent's saved configuration.
	ProviderID         *string  `json:"provider_id,omitempty" jsonschema:"override the agent's provider ID"`
	ModelID            *string  `json:"model_id,omitempty" jsonschema:"override the agent's model ID"`
	SystemPromptID     *string  `json:"system_prompt_id,omitempty" jsonschema:"override the system prompt ID (empty string to clear the prompt)"`
	TemperatureEnabled *bool    `json:"temperature_enabled,omitempty" jsonschema:"override whether temperature sampling is enabled"`
	Temperature        *float64 `json:"temperature,omitempty" jsonschema:"override temperature value"`
	TopPEnabled        *bool    `json:"top_p_enabled,omitempty" jsonschema:"override whether top-p sampling is enabled"`
	TopP               *float64 `json:"top_p,omitempty" jsonschema:"override top-p value"`
	TopKEnabled        *bool    `json:"top_k_enabled,omitempty" jsonschema:"override whether top-k sampling is enabled"`
	TopK               *int     `json:"top_k,omitempty" jsonschema:"override top-k value"`
	MaxToolCalls       *int     `json:"max_tool_calls,omitempty" jsonschema:"override maximum tool call iterations"`

	// Optional tool overrides. If any is provided, the tool selection comes from
	// these fields instead of the agent's saved tool configuration.
	HttpToolIDs    *[]string                   `json:"http_tool_ids,omitempty" jsonschema:"override the HTTP tool IDs for this request"`
	McpServers     *[]McpServerToolConfigInput `json:"mcp_servers,omitempty" jsonschema:"override the MCP server tool selections for this request"`
	BuiltInToolIDs *[]string                   `json:"built_in_tool_ids,omitempty" jsonschema:"override the built-in tool IDs (e.g. web_fetch) for this request"`
}

type PlaygroundChatOutput struct {
	Response  string    `json:"response" jsonschema:"assistant's response"`
	MessageID string    `json:"message_id,omitempty" jsonschema:"ID of the saved assistant message"`
	SessionID string    `json:"session_id" jsonschema:"session the messages were saved to"`
	Role      string    `json:"role" jsonschema:"role of the response message (assistant)"`
	CreatedAt time.Time `json:"created_at,omitempty" jsonschema:"when the assistant message was saved"`
}

// --- Tool registration ---

func (h *Handler) registerPlaygroundTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "playground_sessions_list",
		Description: "List chat sessions for an agent.",
	}, h.listPlaygroundSessions)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "playground_sessions_create",
		Description: "Create a new chat session for an agent. Returns the session id to pass as session_id to playground_chat.",
	}, h.createPlaygroundSession)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "playground_sessions_delete",
		Description: "Delete a chat session.",
	}, h.deletePlaygroundSession)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "playground_messages_list",
		Description: "List all messages in a chat session.",
	}, h.listPlaygroundMessages)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "playground_chat",
		Description: "Send a message to an agent within an existing session (create one with playground_sessions_create) and get a response. The message and response are saved to the session history. Optional fields override the agent's model, sampling and tool settings for this request only. Responses are not streamed.",
	}, h.playgroundChat)
}

// --- Tool handlers ---

func (h *Handler) listPlaygroundSessions(ctx context.Context, _ *mcp.CallToolRequest, input ListPlaygroundSessionsInput) (*mcp.CallToolResult, ListPlaygroundSessionsOutput, error) {
	result, err := h.playgroundApp.Queries.ListSessions.Execute(ctx, queries.ListPlaygroundSessionsParams{
		AgentID: input.AgentID,
	})
	if err != nil {
		return nil, ListPlaygroundSessionsOutput{}, fmt.Errorf("list sessions failed: %w", err)
	}
	sessions := make([]PlaygroundSessionOutput, len(result.Sessions))
	for i, s := range result.Sessions {
		sessions[i] = PlaygroundSessionOutput{
			ID: s.ID, AgentID: s.AgentID, Title: s.Title,
			CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
		}
	}
	return nil, ListPlaygroundSessionsOutput{Sessions: sessions}, nil
}

func (h *Handler) createPlaygroundSession(ctx context.Context, _ *mcp.CallToolRequest, input CreatePlaygroundSessionInput) (*mcp.CallToolResult, PlaygroundSessionOutput, error) {
	result, err := h.playgroundApp.Commands.CreateSession.Execute(ctx, commands.CreatePlaygroundSessionParams{
		AgentID: input.AgentID,
		Title:   input.Title,
	})
	if err != nil {
		return nil, PlaygroundSessionOutput{}, fmt.Errorf("create session failed: %w", err)
	}
	s := result.Session
	return nil, PlaygroundSessionOutput{
		ID: s.ID, AgentID: s.AgentID, Title: s.Title,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}, nil
}

func (h *Handler) deletePlaygroundSession(ctx context.Context, _ *mcp.CallToolRequest, input DeletePlaygroundSessionInput) (*mcp.CallToolResult, DeletePlaygroundSessionOutput, error) {
	err := h.playgroundApp.Commands.DeleteSession.Execute(ctx, commands.DeletePlaygroundSessionParams{
		ID:      input.ID,
		AgentID: input.AgentID,
	})
	if err != nil {
		return nil, DeletePlaygroundSessionOutput{}, fmt.Errorf("delete session failed: %w", err)
	}
	return nil, DeletePlaygroundSessionOutput{Status: "deleted"}, nil
}

func (h *Handler) listPlaygroundMessages(ctx context.Context, _ *mcp.CallToolRequest, input ListPlaygroundMessagesInput) (*mcp.CallToolResult, ListPlaygroundMessagesOutput, error) {
	result, err := h.playgroundApp.Queries.ListMessages.Execute(ctx, queries.ListPlaygroundMessagesParams{
		SessionID: input.SessionID,
		AgentID:   input.AgentID,
	})
	if err != nil {
		return nil, ListPlaygroundMessagesOutput{}, fmt.Errorf("list messages failed: %w", err)
	}
	messages := make([]PlaygroundMessageOutput, len(result.Messages))
	for i, m := range result.Messages {
		messages[i] = PlaygroundMessageOutput{
			ID: m.ID, SessionID: m.SessionID, Role: m.Role,
			Content: m.Content, CreatedAt: m.CreatedAt,
		}
	}
	return nil, ListPlaygroundMessagesOutput{Messages: messages}, nil
}

func (h *Handler) playgroundChat(ctx context.Context, _ *mcp.CallToolRequest, input PlaygroundChatInput) (*mcp.CallToolResult, PlaygroundChatOutput, error) {
	// Save user message
	_, err := h.playgroundApp.Commands.SaveMessage.Execute(ctx, commands.SavePlaygroundMessageParams{
		SessionID: input.SessionID,
		Role:      "user",
		Content:   input.Content,
	})
	if err != nil {
		return nil, PlaygroundChatOutput{}, fmt.Errorf("save user message failed: %w", err)
	}

	// Resolve agent config with tools
	configResult, err := h.playgroundApp.Queries.ResolveConfig.Execute(ctx, toResolvePlaygroundConfigParams(input))
	if err != nil {
		return nil, PlaygroundChatOutput{}, fmt.Errorf("resolve config failed: %w", err)
	}

	// Load conversation history
	messagesResult, err := h.playgroundApp.Queries.ListMessages.Execute(ctx, queries.ListPlaygroundMessagesParams{
		SessionID: input.SessionID,
		AgentID:   input.AgentID,
	})
	if err != nil {
		return nil, PlaygroundChatOutput{}, fmt.Errorf("list messages failed: %w", err)
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

	// Non-streaming chat
	content, err := h.chatProvider.Chat(ctx, chatReq)
	if err != nil {
		return nil, PlaygroundChatOutput{}, fmt.Errorf("chat failed: %w", err)
	}

	out := PlaygroundChatOutput{Response: content, SessionID: input.SessionID, Role: "assistant"}

	// Save assistant response
	if content != "" {
		saved, saveErr := h.playgroundApp.Commands.SaveMessage.Execute(ctx, commands.SavePlaygroundMessageParams{
			SessionID: input.SessionID,
			Role:      "assistant",
			Content:   content,
		})
		if saveErr == nil {
			out.MessageID = saved.Message.ID
			out.CreatedAt = saved.Message.CreatedAt
		}
	}

	return nil, out, nil
}

// toResolvePlaygroundConfigParams maps the chat input, including optional
// overrides, to the app-layer params. Mirrors the HTTP PlaygroundChat handler.
func toResolvePlaygroundConfigParams(input PlaygroundChatInput) queries.ResolvePlaygroundConfigParams {
	params := queries.ResolvePlaygroundConfigParams{
		AgentID:            input.AgentID,
		SystemPromptID:     input.SystemPromptID,
		TemperatureEnabled: input.TemperatureEnabled,
		Temperature:        input.Temperature,
		TopPEnabled:        input.TopPEnabled,
		TopP:               input.TopP,
		TopKEnabled:        input.TopKEnabled,
		TopK:               input.TopK,
		MaxToolCalls:       input.MaxToolCalls,
		IncludeTools:       true,
	}
	if input.ProviderID != nil {
		params.ProviderID = *input.ProviderID
	}
	if input.ModelID != nil {
		params.ModelID = *input.ModelID
	}

	if input.HttpToolIDs != nil || input.McpServers != nil || input.BuiltInToolIDs != nil {
		override := &queries.ToolOverride{}
		if input.HttpToolIDs != nil {
			override.HttpToolIDs = *input.HttpToolIDs
		}
		if input.BuiltInToolIDs != nil {
			override.BuiltInToolIDs = *input.BuiltInToolIDs
		}
		if input.McpServers != nil {
			for _, s := range *input.McpServers {
				toolNames := s.ToolNames
				if toolNames == nil {
					toolNames = []string{}
				}
				override.McpServers = append(override.McpServers, agenttoolrepo.McpServerToolConfig{
					McpServerID: s.McpServerID,
					SelectAll:   s.SelectAll,
					ToolNames:   toolNames,
				})
			}
		}
		params.ToolOverride = override
	}
	return params
}
