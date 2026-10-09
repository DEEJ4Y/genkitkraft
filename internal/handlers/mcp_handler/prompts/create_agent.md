# GenKitKraft — Platform Capabilities

GenKitKraft is a self-hostable LLM agent platform built on [Google Genkit](https://github.com/firebase/genkit) (Go SDK). It ships as a single binary with an embedded frontend and SQLite storage — no external dependencies required.

---

## Table of Contents

1. [Core Concepts](#1-core-concepts)
2. [LLM Providers](#2-llm-providers)
3. [Agent Management](#3-agent-management)
4. [System Prompts](#4-system-prompts)
5. [Tool Integration](#5-tool-integration)
6. [MCP Server (Management via AI Clients)](#6-mcp-server-management-via-ai-clients)
7. [Deploy API (Chat Completions)](#7-deploy-api-chat-completions)
8. [Interactive Playground](#8-interactive-playground)
9. [Authentication & Security](#9-authentication--security)
10. [REST API Reference](#10-rest-api-reference)
11. [Deployment](#11-deployment)
12. [Configuration](#12-configuration)
13. [Creating Agents with GenKitKraft](#13-creating-agents-with-genkitkraft)

---

## 1. Core Concepts

GenKitKraft is built around three primary entities:

- **Providers** — Connections to external LLM services (API key + model selection).
- **Prompts** — Reusable system instruction blocks that define agent behaviour.
- **Agents** — Combinations of a provider, model, prompt, generation parameters, and tools.

Agents are exposed for use through two interfaces: the built-in **Playground** (UI-based testing) and the **Deploy API** (production-grade OpenAI-compatible endpoint).

---

## 2. LLM Providers

GenKitKraft supports multiple LLM providers. Each provider is configured once with an API key, encrypted at rest, and reused across any number of agents. Custom model names can be specified when creating your agent or in the playground.

| Provider      | Available Models                                                                                                                       |
| ------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| **Google AI** | gemini-3.1-pro-preview, gemini-3-flash-preview, gemini-3.1-flash-lite-preview, gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-lite |
| **Vertex AI** | Same models as Google AI (uses Google Cloud credentials instead of an API key)                                                         |
| **OpenAI**    | gpt-5.4, gpt-5.4-mini, gpt-5.4-nano, gpt-5.3-codex, gpt-4o, gpt-4o-mini                                                                |
| **Anthropic** | claude-opus-4-6, claude-sonnet-4-6, claude-opus-4-5-20251202, claude-sonnet-4-5-20250929, claude-haiku-4-5-20251001                    |
| **xAI**       | grok-4.20-beta, grok-4-1-fast-reasoning, grok-4-1-fast-non-reasoning, grok-4, grok-code-fast-1                                         |
| **DeepSeek**  | deepseek-chat, deepseek-reasoner                                                                                                       |

The model list changes over time. This list is a guide only. Use `provider_types_list` to see the supported provider types and their `model_prefix` values. You can type a custom model name when you create an agent.

### Provider Operations

- **Add** — Select provider type, enter display name and API key.
- **Test** — Validate API key and reachability with a single click.
- **Edit** — Update display name or rotate the API key.
- **Delete** — Remove a provider. Agents using it must be reconfigured.

All API keys are encrypted at rest using the `ENCRYPTION_KEY` environment variable.

---

## 3. Agent Management

Agents are the central unit. Each agent packages a provider, model, system prompt, generation parameters, and tool assignments.

### Agent Configuration Fields

| Field             | Description                                            |
| ----------------- | ------------------------------------------------------ |
| **Name**          | Display name for the agent                             |
| **Provider**      | Which configured LLM provider to use                   |
| **Model**         | Model from the selected provider                       |
| **System Prompt** | A saved prompt from the Prompts library                |
| **Temperature**   | Controls randomness (0.0–2.0)                          |
| **Top P**         | Nucleus sampling threshold                             |
| **Top K**         | Limits the token vocabulary per step                   |
| **Max Tool Calls** | Maximum number of tool call iterations per request (`max_tool_calls`, default 10) |
| **Gap Reporting** | Lets the agent report gaps with the `report_gap` tool (`gap_reporting_enabled`, default false). See [Gaps](#31-gaps) |
| **Tools**         | HTTP tools, MCP server tools and built-in tools (e.g. web_fetch) assigned to this agent |

Generation parameters (temperature, top P, top K) are optional — omitting them uses provider defaults. Each one has an `*_enabled` flag (`temperature_enabled`, `top_p_enabled`, `top_k_enabled`). Set the flag to `true` to send the value to the provider.

### Agent Operations

- Create, read, update, delete agents via UI or API.
- Assign and re-assign tools independently of the agent definition.
- Use any agent immediately in the Playground or via the Deploy API.

### 3.1 Gaps

An agent with Gap Reporting on gets a `report_gap` tool. The agent calls it during a conversation when it notices a gap. The call does not change the answer that the user receives. A gap has one of three categories:

- **knowledge** — the agent could not answer a question reliably.
- **capability** — the agent could not do a requested action (a missing tool, permission, or integration).
- **improvement** — nothing failed, but the agent sees a way to automate more of the flow.

A background process compares each new report with the open gaps of the agent. It merges the report into a matching gap or creates a new gap. Gaps belong to one agent. `report_gap` is not a tool that you assign. It is on when `gap_reporting_enabled` is `true`.

Gap status values: `open`, `resolved`, `dismissed`. Use these MCP tools to review gaps:

- `gaps_list` — list the gaps of an agent (filter with `status`).
- `gaps_get` — get one gap, with the sessions and messages where it was seen.
- `gaps_resolve` — mark a gap as resolved after you fix the cause.
- `gaps_dismiss` — dismiss a gap with `dismissal_category`: `unrelated`, `insufficient_detail`, `duplicate`, or `other`. A gap dismissed as `unrelated` is permanent and you cannot reopen it.
- `gaps_reopen` — reopen a resolved or dismissed gap. A new report that matches a resolved or dismissed gap also reopens it. The earlier dismissal stays as history.

---

## 4. System Prompts

Prompts are reusable system instructions stored in a library and referenced by agents. Decoupling prompts from agents lets you share a single prompt across multiple agents or swap prompts without recreating agents.

### Prompt Operations

- **Create** — Name and write the system instruction content.
- **Edit** — Update name or content at any time.
- **Delete** — Remove a prompt. Agents that reference it must be updated.

### Tips for Writing System Prompts

- Clearly state the agent's role and purpose.
- Include constraints on tone, format, or domain.
- Use explicit instructions rather than implicit assumptions.

---

## 5. Tool Integration

Agents can call external tools during inference. GenKitKraft supports three tool types: HTTP tools, MCP servers, and built-in tools.

### 5.1 HTTP Tools

Point-and-shoot integration with any HTTP endpoint. Define once, reuse across agents.

| Field            | Description                                                                  |
| ---------------- | ---------------------------------------------------------------------------- |
| **Name**         | Tool identifier shown to the LLM                                             |
| **Description**  | Natural-language description used by the LLM to decide when to call the tool |
| **Method**       | HTTP method: GET, POST, PUT, DELETE, or PATCH                                |
| **URL**          | Target endpoint                                                              |
| **Headers**      | Static key/value headers (e.g., auth tokens)                                 |
| **Body**         | Optional static or templated request body                                    |
| **Input Schema** | JSON Schema describing the tool's input parameters                           |

When running in Docker, use the internal Docker network hostname instead of `localhost` for tool URLs.

### 5.2 MCP Servers

Connect external MCP (Model Context Protocol) tool servers. GenKitKraft acts as an MCP client and proxies tool calls to registered servers.

| Field         | Description                |
| ------------- | -------------------------- |
| **Name**      | Display name               |
| **Transport** | `sse` or `streamableHttp` |
| **URL**       | MCP server endpoint        |
| **Headers**   | Optional auth headers      |

**Transport auto-fallback**: If `streamableHttp` is selected and the server doesn't support it, GenKitKraft automatically falls back to SSE transport.

After registering an MCP server, use **Discover Tools** to fetch the list of available tools from that server.

### 5.3 Built-in Tools

Built-in tools are part of GenKitKraft. You do not define them. Use `built_in_tools_list` to see the available tools and their IDs. Today the list has `web_fetch`. It fetches static content from a URL and returns it as Markdown. It does not run JavaScript.

### 5.4 Agent Tool Assignment

Each agent has an independent tool configuration:

- Select any combination of HTTP tools (`http_tool_ids`).
- Select specific tools from registered MCP servers (`mcp_servers`). Not all tools from a server need to be enabled.
- Select built-in tools (`built_in_tool_ids`).
- Tool configuration can be updated without modifying the agent itself.

> ⚠️ `agent_tools_update` replaces the whole configuration. Call `agent_tools_get` first. Then send the current values of everything you want to keep in `http_tool_ids`, `mcp_servers`, and `built_in_tool_ids`. If you omit `built_in_tool_ids`, the call disables all built-in tools.

### 5.5 Playground Tool Overrides

In the Playground, tool assignments can be overridden per session without affecting the saved agent configuration. Overrides can be promoted back to the saved config with **Save Configuration**. Over MCP, `playground_chat` accepts `http_tool_ids`, `mcp_servers`, and `built_in_tool_ids` as overrides for one request.

---

## 6. MCP Server (Management via AI Clients)

GenKitKraft exposes its entire management API as MCP tools, allowing you to create and manage agents conversationally from Claude Desktop, Cursor, or any MCP-compatible client.

### Endpoint

```
http://<host>:<port>/mcp
```

Default: `http://localhost:8080/mcp`  
Transport: **Streamable HTTP**

### Authentication

When `AUTH_CREDENTIALS` is set, the MCP endpoint requires HTTP Basic Auth with the same credentials.

Claude Desktop (uses the `mcp-remote` bridge):

```json
{
  "mcpServers": {
    "genkitkraft": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "http://localhost:8080/mcp",
        "--header",
        "Authorization:${AUTH_HEADER}"
      ],
      "env": {
        "AUTH_HEADER": "Basic <base64(username:password)>"
      }
    }
  }
}
```

Cursor and other clients that connect by URL:

```json
{
  "mcpServers": {
    "genkitkraft": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Basic <base64(username:password)>"
      }
    }
  }
}
```

Generate the header value: `echo -n "admin:changeme" | base64`

### Available MCP Tools

#### Agents

| Tool            | Description                     |
| --------------- | ------------------------------- |
| `agents_list`   | List all agents with pagination |
| `agents_get`    | Get an agent by ID              |
| `agents_create` | Create a new agent              |
| `agents_update` | Update an agent                 |
| `agents_delete` | Delete an agent                 |

#### Agent Tool Configuration

| Tool                 | Description                              |
| -------------------- | ---------------------------------------- |
| `agent_tools_get`    | Get which tools are assigned to an agent |
| `agent_tools_update` | Update tool assignments for an agent     |

#### Prompts

| Tool             | Description                |
| ---------------- | -------------------------- |
| `prompts_list`   | List all system prompts    |
| `prompts_get`    | Get a prompt by ID         |
| `prompts_create` | Create a new system prompt |
| `prompts_update` | Update a prompt            |
| `prompts_delete` | Delete a prompt            |

#### Providers

| Tool                  | Description                   |
| --------------------- | ----------------------------- |
| `providers_list`      | List all LLM providers        |
| `providers_get`       | Get a provider by ID          |
| `providers_create`    | Create a new provider         |
| `providers_update`    | Update a provider             |
| `providers_delete`    | Delete a provider             |
| `providers_test`      | Test provider connectivity    |
| `provider_types_list` | List supported provider types |

#### HTTP Tools

| Tool                | Description            |
| ------------------- | ---------------------- |
| `http_tools_list`   | List all HTTP tools    |
| `http_tools_get`    | Get an HTTP tool by ID |
| `http_tools_create` | Create a new HTTP tool |
| `http_tools_update` | Update an HTTP tool    |
| `http_tools_delete` | Delete an HTTP tool    |

#### MCP Servers

| Tool                     | Description                       |
| ------------------------ | --------------------------------- |
| `mcp_servers_list`       | List all registered MCP servers   |
| `mcp_servers_get`        | Get an MCP server by ID           |
| `mcp_servers_create`     | Register a new MCP server         |
| `mcp_servers_update`     | Update an MCP server              |
| `mcp_servers_delete`     | Delete an MCP server              |
| `mcp_servers_list_tools` | Discover tools from an MCP server |

#### Playground

| Tool                         | Description                       |
| ---------------------------- | --------------------------------- |
| `playground_sessions_list`   | List chat sessions for an agent   |
| `playground_sessions_create` | Create a new chat session         |
| `playground_sessions_delete` | Delete a chat session             |
| `playground_messages_list`   | List messages in a session        |
| `playground_chat`            | Send a message and get a response |

Send the session ID as `session_id`. If you cannot send `session_id`, send the same value as `fallback_session_id`. It takes precedence over `session_id`. In `playground_chat`, the message text goes in `content`. `playground_chat` also accepts optional overrides for one request: `provider_id`, `model_id`, `system_prompt_id`, the sampling fields, `max_tool_calls`, and the tool fields. The overrides do not change the saved agent.

#### Gaps

| Tool            | Description                                                |
| --------------- | ---------------------------------------------------------- |
| `gaps_list`     | List the gaps an agent reported, with pagination           |
| `gaps_get`      | Get a gap by ID                                            |
| `gaps_resolve`  | Mark a gap as resolved                                     |
| `gaps_dismiss`  | Dismiss a gap with a reason category                       |
| `gaps_reopen`   | Reopen a resolved or dismissed gap                         |

#### Built-in Tools

| Tool                  | Description                                                               |
| --------------------- | ------------------------------------------------------------------------- |
| `built_in_tools_list` | List all built-in tools. Assign them with `agent_tools_update` (`built_in_tool_ids`) |

#### Auth

| Tool              | Description                   |
| ----------------- | ----------------------------- |
| `auth_login`      | Log in with username/password |
| `auth_logout`     | Log out                       |
| `auth_get_me`     | Get current user info         |
| `auth_get_status` | Check if auth is required     |

#### Health

| Tool               | Description     |
| ------------------ | --------------- |
| `health_liveness`  | Liveness check  |
| `health_readiness` | Readiness check |

### Built-in `create-agent` Prompt

The MCP server ships with a `create-agent` server-side prompt. MCP clients that support server-side prompts (like Claude Desktop) can load this to give the LLM a full walkthrough of the correct agent-creation workflow. The server also has a `backup` prompt and a `restore` prompt for GenKitKraft configurations.

### Example Agent Creation Workflow via MCP

1. `provider_types_list` — see available provider types
2. `providers_create` — configure an LLM provider
3. `prompts_create` — write a system prompt
4. `agents_create` — create the agent (optional: `max_tool_calls`, `gap_reporting_enabled`)
5. `http_tools_create`, `mcp_servers_create`, `built_in_tools_list` — find or create the tools the agent needs
6. `agent_tools_get`, then `agent_tools_update` — assign the tools. Send the full configuration, including `built_in_tool_ids`
7. `playground_sessions_create` — start a chat session
8. `playground_chat` — test the agent
9. `gaps_list` — if Gap Reporting is on, review the gaps that the agent reported

---

## 7. Deploy API (Chat Completions)

Configured agents are exposed through an **OpenAI-compatible** chat completions endpoint. Any application or library built for the OpenAI API works out of the box.

### Authentication

Controlled by the `PUBLIC_API_KEY` environment variable.

- When set: requests must include `Authorization: Bearer <key>`
- When unset: deploy endpoints are publicly accessible

This is separate from the `AUTH_CREDENTIALS` used for the management UI.

### 7.1 Stateless Chat Completions

The caller provides the full message history on every request.

```
POST /api/v1/agents/{agentId}/deploy/chat/completions
```

#### Request

```json
{
  "messages": [{ "role": "user", "content": "Hello!" }],
  "stream": false
}
```

- `messages` — array of `user`/`assistant` messages. System messages are not supported (the agent's system prompt is applied automatically).
- `stream` — set to `true` for Server-Sent Events streaming.

#### Response (non-streaming)

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1700000000,
  "model": "my-agent",
  "choices": [
    {
      "index": 0,
      "message": { "role": "assistant", "content": "Hello! I'm..." },
      "finish_reason": "stop"
    }
  ],
  "usage": { "prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0 }
}
```

#### Response (streaming — SSE)

```
data: {"id":"...","object":"chat.completion.chunk","choices":[{"delta":{"role":"assistant"}}]}
data: {"id":"...","object":"chat.completion.chunk","choices":[{"delta":{"content":"Hello"}}]}
data: [DONE]
```

### 7.2 Stateful Chat (Sessions)

The server manages conversation history. The caller only sends the new message each turn.

#### Session Lifecycle

```
POST   /api/v1/agents/{agentId}/deploy/sessions               → create
GET    /api/v1/agents/{agentId}/deploy/sessions               → list (paginated: ?limit=&offset=)
GET    /api/v1/agents/{agentId}/deploy/sessions/{sessionId}   → get
GET    /api/v1/agents/{agentId}/deploy/sessions/{sessionId}/messages → conversation history
DELETE /api/v1/agents/{agentId}/deploy/sessions/{sessionId}   → delete (clears all messages)
```

#### Stateful Chat Request

```
POST /api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions
```

Only the last user message in `messages` is used. Full history is loaded from the session automatically and both the new message and the response are persisted.

#### Reconnect and Cancel

```
GET  /api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions/stream  → reconnect to the reply that is streaming (SSE)
POST /api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions/cancel  → stop the reply that is streaming (204)
```

On reconnect, send the `Last-Event-ID` header with the last SSE `id:` that you received. The stream then continues from the next token. Cancel is best-effort. It does nothing if the reply is complete.

### 7.3 Client Examples

#### Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/api/v1/agents/{agentId}/deploy",
    api_key="my-secret-key",
)
response = client.chat.completions.create(
    model="any",
    messages=[{"role": "user", "content": "Hello!"}],
)
```

#### Node.js (OpenAI SDK)

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "http://localhost:8080/api/v1/agents/{agentId}/deploy",
  apiKey: "my-secret-key",
});
const response = await client.chat.completions.create({
  model: "any",
  messages: [{ role: "user", content: "Hello!" }],
});
```

### 7.4 Error Format

Errors follow the OpenAI error format:

```json
{
  "error": {
    "message": "messages must not be empty",
    "type": "invalid_request_error",
    "code": "invalid_request"
  }
}
```

| Status | Reason                                                            |
| ------ | ----------------------------------------------------------------- |
| 400    | Invalid request (empty messages, system messages, malformed JSON) |
| 401    | Missing or invalid API key                                        |
| 404    | Agent not found                                                   |
| 500    | Internal server error                                             |

---

## 8. Interactive Playground

The Playground provides a browser-based chat interface for testing agents directly. It streams responses in real time via SSE.

### Features

- **Session management** — Create, switch between, and delete named sessions. History is preserved per session.
- **Configuration overrides** — Override the agent's provider, model, system prompt, temperature, top P, top K, or max tool calls for the current session without modifying the saved agent.
- **Tool overrides** — Enable/disable specific HTTP tools, MCP server tools, or built-in tools for the session.
- **Save configuration** — Promote playground overrides back to the saved agent configuration.
- **Streaming responses** — Output appears token by token. A client can reconnect to a reply that is streaming, and can cancel it.
- **Gap reporting** — If the agent has Gap Reporting on, gaps reported in the Playground appear on the Gaps tab of the agent.

### Playground API

```
GET  /api/v1/agents/{agentId}/playground/sessions             → list sessions
POST /api/v1/agents/{agentId}/playground/sessions             → create session
DELETE /api/v1/agents/{agentId}/playground/sessions/{id}      → delete session
GET  /api/v1/agents/{agentId}/playground/sessions/{id}/messages → list messages
POST /api/v1/agents/{agentId}/playground/chat                 → send message (SSE by default, or JSON with "stream": false)
GET  /api/v1/agents/{agentId}/playground/sessions/{id}/stream → reconnect to the reply that is streaming (SSE)
POST /api/v1/agents/{agentId}/playground/sessions/{id}/stream/cancel → stop the reply that is streaming (204)
```

### Chat Widget

You can put an agent on a website with the open-source AI chat widget. The widget builder (in the dashboard under **Agents → your agent → Widget**) designs the look and gives the embed code. The widget must call your own backend. Your backend then calls the Deploy API with the `PUBLIC_API_KEY`. Never put the key in browser code. For the full guide, see the `Embedding an Agent with the Chat Widget` page in the GenKitKraft documentation.

---

## 9. Authentication & Security

### Management UI / API Authentication

Controlled by `AUTH_CREDENTIALS`. When set:

- All UI and management API access requires login.
- Users log in with username + password.
- A `HttpOnly` session cookie is issued on success.
- The same credentials apply to the MCP endpoint (Basic Auth).
- The login endpoint is rate-limited to prevent brute-force attacks.

Multiple users can be configured: `AUTH_CREDENTIALS=admin:password1,user2:password2`

When `AUTH_CREDENTIALS` is unset, all access is open.

### Deploy API Authentication

Controlled by `PUBLIC_API_KEY` (separate from `AUTH_CREDENTIALS`).

- Multiple keys supported: `PUBLIC_API_KEY=sk-key-one,sk-key-two`
- Clients send `Authorization: Bearer <key>` on deploy requests.
- When unset, deploy endpoints are public.

### Encryption at Rest

All LLM provider API keys are encrypted in the SQLite database using the `ENCRYPTION_KEY` secret. The server refuses to start if this variable is missing.

### Auth Endpoints

| Endpoint           | Method | Description                         |
| ------------------ | ------ | ----------------------------------- |
| `/api/auth/status` | GET    | Check if authentication is required |
| `/api/auth/login`  | POST   | Log in (returns session cookie)     |
| `/api/auth/logout` | POST   | Log out (clears session)            |
| `/api/auth/me`     | GET    | Get current authenticated user      |

---

## 10. REST API Reference

All endpoints are served on the same port as the UI (default: `8080`).

### Health Probes

| Method | Path      | Description                                  |
| ------ | --------- | -------------------------------------------- |
| GET    | `/livez`  | Liveness — returns 200 if server is running  |
| GET    | `/readyz` | Readiness — returns 200 if ready, 503 if not |

### Provider Types

| Method | Path                     | Description                                 |
| ------ | ------------------------ | ------------------------------------------- |
| GET    | `/api/v1/provider-types` | List supported provider types with metadata |

### Providers

| Method | Path                                   | Description                   |
| ------ | -------------------------------------- | ----------------------------- |
| GET    | `/api/v1/settings/providers`           | List all configured providers |
| POST   | `/api/v1/settings/providers`           | Create a new provider         |
| GET    | `/api/v1/settings/providers/{id}`      | Get a specific provider       |
| PUT    | `/api/v1/settings/providers/{id}`      | Update a provider             |
| DELETE | `/api/v1/settings/providers/{id}`      | Delete a provider             |
| POST   | `/api/v1/settings/providers/{id}/test` | Test provider connectivity    |

### Prompts

| Method | Path                   | Description                      |
| ------ | ---------------------- | -------------------------------- |
| GET    | `/api/v1/prompts`      | List prompts (`?limit=&offset=`) |
| POST   | `/api/v1/prompts`      | Create a new prompt              |
| GET    | `/api/v1/prompts/{id}` | Get a specific prompt            |
| PUT    | `/api/v1/prompts/{id}` | Update a prompt                  |
| DELETE | `/api/v1/prompts/{id}` | Delete a prompt                  |

### Agents

| Method | Path                  | Description                     |
| ------ | --------------------- | ------------------------------- |
| GET    | `/api/v1/agents`      | List agents (`?limit=&offset=`) |
| POST   | `/api/v1/agents`      | Create a new agent              |
| GET    | `/api/v1/agents/{id}` | Get a specific agent            |
| PUT    | `/api/v1/agents/{id}` | Update an agent                 |
| DELETE | `/api/v1/agents/{id}` | Delete an agent                 |

### Agent Tools

| Method | Path                             | Description                       |
| ------ | -------------------------------- | --------------------------------- |
| GET    | `/api/v1/agents/{agentId}/tools` | Get agent's tool configuration    |
| PUT    | `/api/v1/agents/{agentId}/tools` | Update agent's tool configuration |

### HTTP Tools

| Method | Path                      | Description                         |
| ------ | ------------------------- | ----------------------------------- |
| GET    | `/api/v1/http-tools`      | List HTTP tools (`?limit=&offset=`) |
| POST   | `/api/v1/http-tools`      | Create an HTTP tool                 |
| GET    | `/api/v1/http-tools/{id}` | Get a specific HTTP tool            |
| PUT    | `/api/v1/http-tools/{id}` | Update an HTTP tool                 |
| DELETE | `/api/v1/http-tools/{id}` | Delete an HTTP tool                 |

### MCP Servers

| Method | Path                             | Description                          |
| ------ | -------------------------------- | ------------------------------------ |
| GET    | `/api/v1/mcp-servers`            | List MCP servers (`?limit=&offset=`) |
| POST   | `/api/v1/mcp-servers`            | Create an MCP server                 |
| GET    | `/api/v1/mcp-servers/{id}`       | Get a specific MCP server            |
| PUT    | `/api/v1/mcp-servers/{id}`       | Update an MCP server                 |
| DELETE | `/api/v1/mcp-servers/{id}`       | Delete an MCP server                 |
| GET    | `/api/v1/mcp-servers/{id}/tools` | Discover tools from an MCP server    |

### Deploy (Chat Completions)

| Method | Path                                                                    | Description                 |
| ------ | ----------------------------------------------------------------------- | --------------------------- |
| POST   | `/api/v1/agents/{agentId}/deploy/chat/completions`                      | Stateless chat completions  |
| POST   | `/api/v1/agents/{agentId}/deploy/sessions`                              | Create a stateful session   |
| GET    | `/api/v1/agents/{agentId}/deploy/sessions`                              | List sessions (paginated)   |
| GET    | `/api/v1/agents/{agentId}/deploy/sessions/{sessionId}`                  | Get session metadata        |
| GET    | `/api/v1/agents/{agentId}/deploy/sessions/{sessionId}/messages`         | Get conversation history    |
| DELETE | `/api/v1/agents/{agentId}/deploy/sessions/{sessionId}`                  | Delete session and messages |
| POST   | `/api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions` | Stateful chat completions   |

### Playground

| Method | Path                                                         | Description          |
| ------ | ------------------------------------------------------------ | -------------------- |
| GET    | `/api/v1/agents/{agentId}/playground/sessions`               | List sessions        |
| POST   | `/api/v1/agents/{agentId}/playground/sessions`               | Create session       |
| DELETE | `/api/v1/agents/{agentId}/playground/sessions/{id}`          | Delete session       |
| GET    | `/api/v1/agents/{agentId}/playground/sessions/{id}/messages` | Get messages         |
| POST   | `/api/v1/agents/{agentId}/playground/chat`                   | Chat (SSE streaming) |

### Pagination

All list endpoints support:

```
?limit=20&offset=0
```

### Response Format

Success responses return the resource directly as JSON. Errors return:

```json
{ "message": "Description of the error" }
```

---

## 11. Deployment

### Docker (Recommended)

```bash
docker pull ghcr.io/deej4y/genkitkraft:latest

docker run -d \
  -p 8080:8080 \
  -v genkitkraft-data:/data \
  -e ENCRYPTION_KEY=$(openssl rand -base64 32) \
  ghcr.io/deej4y/genkitkraft:latest
```

### Docker Compose

```yaml
services:
  genkitkraft:
    image: ghcr.io/deej4y/genkitkraft:latest
    ports:
      - "8080:8080"
    volumes:
      - genkitkraft-data:/data
    environment:
      PORT: 8080
      DATABASE_PATH: /data/app.db
      ENCRYPTION_KEY: your-encryption-key
      # AUTH_CREDENTIALS: admin:changeme
      # PUBLIC_API_KEY: sk-my-secret-key

volumes:
  genkitkraft-data:
```

### From Source

**Prerequisites:** Go 1.26+, Node.js 22+

```bash
cd ui && npm ci && npm run build && cd ..
go build ./cmd/server/...
ENCRYPTION_KEY=my-secret-key PORT=8080 ./server
```

### Reverse Proxy (Caddy)

```caddyfile
yourdomain.com {
    reverse_proxy localhost:8080
}
```

### Reverse Proxy (Nginx)

Requires SSE-specific headers for streaming to work:

```nginx
location / {
    proxy_pass http://localhost:8080;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_buffering off;           # required for SSE
    proxy_cache off;
    proxy_set_header Connection '';
    proxy_http_version 1.1;
    chunked_transfer_encoding on;
}
```

### Health Checks

Both probes are available for container orchestrators:

- `GET /livez` — liveness
- `GET /readyz` — readiness (503 until fully ready)

---

## 12. Configuration

GenKitKraft is configured entirely through environment variables. No config files are needed.

| Variable           | Description                                               | Default                      | Required |
| ------------------ | --------------------------------------------------------- | ---------------------------- | -------- |
| `PORT`             | HTTP server port                                          | `8080`                       | No       |
| `DATABASE_PATH`    | Path to SQLite database file                              | `/data/app.db`               | No       |
| `ENCRYPTION_KEY`   | Secret key for encrypting provider API keys at rest       | —                            | **Yes**  |
| `AUTH_CREDENTIALS` | Comma-separated `username:password` pairs for UI/API auth | _(unset — auth disabled)_    | No       |
| `PUBLIC_API_KEY`   | Comma-separated API keys for deploy endpoints             | _(unset — deploy is public)_ | No       |

### Notes

- **`ENCRYPTION_KEY`** — The server will not start without this. Generate a strong key with `openssl rand -base64 32`. If lost, all saved provider configurations become unreadable.
- **`DATABASE_PATH`** — In Docker, point this at a persistent volume path to survive container recreation.
- **`AUTH_CREDENTIALS`** — Also gates the MCP endpoint (Basic Auth). Format: `admin:pass1,user2:pass2`.
- **`PUBLIC_API_KEY`** — Independent of `AUTH_CREDENTIALS`. Controls access to the Deploy API only. Multiple keys: `sk-key-one,sk-key-two`.

## 13. Creating Agents with GenKitKraft

Agents are created by combining a configured LLM provider, a system prompt, generation parameters, and tool assignments. This modular design allows you to reuse providers and prompts across multiple agents.

### Steps to Create an Agent

1. **Configure an LLM Provider** — Add your API key and select the model you want to use.
2. **Create a System Prompt** — Write the instructions that will guide the agent's behaviour.
3. **Define the Agent** — Give it a name, select the provider and prompt, and optionally set any generation parameters (temperature, top P, top K), the maximum number of tool calls, and Gap Reporting.
4. **Assign Tools** — Choose which HTTP tools, MCP server tools and built-in tools (such as web_fetch) the agent can use during inference.
5. **Save and Deploy** — Once saved, the agent is immediately available for testing in the Playground or via the Deploy API.
6. **Test and Iterate** — Use the Playground to have conversations with your agent, tweak the system prompt, adjust generation parameters, or reassign tools as needed.
7. **Review Gaps** — If Gap Reporting is on, check the Gaps tab (or `gaps_list`) for knowledge, capability, and improvement gaps. Fix the cause, then resolve or dismiss each gap.
8. **Embed (optional)** — Use the chat widget to put the agent on a website.

### Tips for Effective Agents

- Start with a clear, extensively detailed, and specific system prompt to define the agent's role and behaviour.
- Use tools to augment the agent's capabilities, but be mindful of overloading it with too many options.
- Set `max_tool_calls` high enough for the multi-step tasks that you expect, but low enough to stop runaway tool loops.
- Leverage the Playground for rapid iteration and testing before finalizing your agent's configuration for production use.
