---
sidebar_position: 2
---

# Endpoint Reference

Complete reference of all GenKitKraft API endpoints.

## Health

| Method | Path | Description |
|---|---|---|
| GET | `/livez` | Liveness probe — returns 200 if the server is running |
| GET | `/readyz` | Readiness probe — returns 200 if ready, 503 if not |

## Authentication

| Method | Path | Description |
|---|---|---|
| GET | `/api/auth/status` | Check if authentication is required |
| POST | `/api/auth/login` | Log in with username and password |
| POST | `/api/auth/logout` | Log out and clear session |
| GET | `/api/auth/me` | Get current authenticated user |

### POST /api/auth/login

Request:
```json
{"username": "string", "password": "string"}
```

Responses: 200 (success, sets cookie), 401 (invalid credentials), 429 (rate limited)

## Provider Types

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/provider-types` | List supported provider types with metadata |

## Providers

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/settings/providers` | List all configured providers |
| POST | `/api/v1/settings/providers` | Create a new provider |
| GET | `/api/v1/settings/providers/{id}` | Get a specific provider |
| PUT | `/api/v1/settings/providers/{id}` | Update a provider |
| DELETE | `/api/v1/settings/providers/{id}` | Delete a provider |
| POST | `/api/v1/settings/providers/{id}/test` | Test provider connectivity |

### POST /api/v1/settings/providers

Request:
```json
{
  "name": "string",
  "providerType": "string",
  "apiKey": "string"
}
```
Responses: 201 (created), 409 (conflict)

### PUT /api/v1/settings/providers/\{id\}

Request:
```json
{
  "name": "string",
  "apiKey": "string"
}
```
Responses: 200 (updated), 404 (not found), 409 (conflict)

## Prompts

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/prompts` | List prompts (supports `limit` and `offset`) |
| POST | `/api/v1/prompts` | Create a new prompt |
| GET | `/api/v1/prompts/{id}` | Get a specific prompt |
| PUT | `/api/v1/prompts/{id}` | Update a prompt |
| DELETE | `/api/v1/prompts/{id}` | Delete a prompt |

### POST /api/v1/prompts

Request:
```json
{
  "name": "string",
  "content": "string"
}
```
Responses: 201 (created)

## Agents

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/agents` | List agents (supports `limit` and `offset`) |
| POST | `/api/v1/agents` | Create a new agent |
| GET | `/api/v1/agents/{id}` | Get a specific agent |
| PUT | `/api/v1/agents/{id}` | Update an agent |
| DELETE | `/api/v1/agents/{id}` | Delete an agent |

### POST /api/v1/agents

Request:
```json
{
  "name": "string",
  "providerId": "string",
  "model": "string",
  "promptId": "string",
  "generationConfig": {
    "temperature": 0.7,
    "topP": 0.9,
    "topK": 40
  }
}
```
Responses: 201 (created)

## Deploy (Chat Completions)

| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/agents/{agentId}/deploy/chat/completions` | Stateless chat completions (caller provides full history) |
| POST | `/api/v1/agents/{agentId}/deploy/sessions` | Create a new stateful chat session |
| GET | `/api/v1/agents/{agentId}/deploy/sessions/{sessionId}` | Get session metadata |
| DELETE | `/api/v1/agents/{agentId}/deploy/sessions/{sessionId}` | Delete a session and all its messages |
| POST | `/api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions` | Stateful chat completions (server manages history) |

See the full [Deploy API documentation](./deploy) for request/response format, authentication, and examples.

## Playground

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/agents/{agentId}/playground/sessions` | List sessions for an agent |
| POST | `/api/v1/agents/{agentId}/playground/sessions` | Create a new session |
| DELETE | `/api/v1/agents/{agentId}/playground/sessions/{sessionId}` | Delete a session |
| GET | `/api/v1/agents/{agentId}/playground/sessions/{sessionId}/messages` | Get messages in a session |
| POST | `/api/v1/agents/{agentId}/playground/chat` | Chat with an agent (SSE streaming) |

### POST /api/v1/agents/\{agentId\}/playground/chat

Request:
```json
{
  "message": "string",
  "sessionId": "string",
  "configOverrides": {
    "providerId": "string",
    "model": "string",
    "temperature": 0.7,
    "topP": 0.9,
    "topK": 40
  }
}
```

Response: Server-Sent Events (SSE) stream with content chunks.
