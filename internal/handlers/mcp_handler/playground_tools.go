package mcphandler

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
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
	Title   string `json:"title" jsonschema:"session title (required)"`
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

type PlaygroundChatInput struct {
	AgentID   string `json:"agent_id" jsonschema:"agent ID (required)"`
	SessionID string `json:"session_id" jsonschema:"session ID (required)"`
	Content   string `json:"content" jsonschema:"user message content (required)"`
}

type PlaygroundChatOutput struct {
	Response string `json:"response" jsonschema:"assistant's response"`
}

// --- Tool registration ---

func (h *Handler) registerPlaygroundTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "playground_sessions_list",
		Description: "List chat sessions for an agent.",
	}, h.listPlaygroundSessions)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "playground_sessions_create",
		Description: "Create a new chat session for an agent.",
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
		Description: "Send a message to an agent and get a response. The message and response are saved to the session history.",
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

func (h *Handler) deletePlaygroundSession(ctx context.Context, _ *mcp.CallToolRequest, input DeletePlaygroundSessionInput) (*mcp.CallToolResult, any, error) {
	err := h.playgroundApp.Commands.DeleteSession.Execute(ctx, commands.DeletePlaygroundSessionParams{
		ID:      input.ID,
		AgentID: input.AgentID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("delete session failed: %w", err)
	}
	return nil, map[string]string{"status": "deleted"}, nil
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
	configResult, err := h.playgroundApp.Queries.ResolveConfig.Execute(ctx, queries.ResolvePlaygroundConfigParams{
		AgentID:      input.AgentID,
		IncludeTools: true,
	})
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

	// Save assistant response
	if content != "" {
		_, _ = h.playgroundApp.Commands.SaveMessage.Execute(ctx, commands.SavePlaygroundMessageParams{
			SessionID: input.SessionID,
			Role:      "assistant",
			Content:   content,
		})
	}

	return nil, PlaygroundChatOutput{Response: content}, nil
}
