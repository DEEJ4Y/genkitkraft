# Creating Agents via MCP

This guide explains how to create and configure GenKitKraft agents using the MCP (Model Context Protocol) tools. You can follow these steps manually through any MCP client, or paste the prompt below into an LLM connected to GenKitKraft's MCP server and let it handle everything.

## Prerequisites

- GenKitKraft running (see [Docker setup](/docs/guides/docker))
- MCP client connected to GenKitKraft (see [MCP Quickstart](/docs/guides/mcp-quickstart))

## The `create-agent` Prompt

GenKitKraft's MCP server includes a built-in **`create-agent` prompt** that provides the LLM with a comprehensive guide to all available tools and the correct order to use them. MCP clients that support server-side prompts (like Claude Desktop) can load this prompt automatically.

If your MCP client doesn't support prompts, you can copy the guide below and paste it into your LLM's context.

## Agent Creation Workflow

Follow these steps in order. Each step produces an ID needed by the next.

```
provider_types_list → pick a provider type
        ↓
providers_create → provider_id
        ↓
prompts_create → prompt_id (optional)
        ↓
agents_create → agent_id
        ↓
[http_tools_create / mcp_servers_create → tool IDs] (optional)
        ↓
[agent_tools_update → assign tools to agent] (optional)
        ↓
playground_sessions_create → session_id
        ↓
playground_chat → agent response
```

### 1. List provider types

Call `provider_types_list` to see supported LLM providers (openai, google_ai, anthropic, etc.) and what each one requires.

### 2. Create a provider

Call `providers_create` with:
- `name` — a display name
- `provider_type` — from step 1 (e.g., `openai`)
- `api_key` — your API key for the provider

Optionally set `base_url` for custom endpoints (Azure, local models).

Save the returned **`id`** as your `provider_id`.

Use `providers_test` with the provider ID to verify connectivity.

### 3. Create a system prompt (optional)

Call `prompts_create` with:
- `name` — prompt name
- `content` — the system instructions for the agent

Save the returned **`id`** as your `prompt_id`.

### 4. Create the agent

Call `agents_create` with:
- `name` — agent name
- `provider_id` — from step 2
- `model_id` — the model to use (e.g., `gpt-4o`, `gemini-2.0-flash`, `claude-sonnet-4-20250514`)
- `system_prompt_id` — from step 3 (optional)

Save the returned **`id`** as your `agent_id`.

### 5. Assign tools (optional)

If you want your agent to call external APIs or MCP servers:

1. Create HTTP tools with `http_tools_create` and/or register external MCP servers with `mcp_servers_create`
2. Discover MCP server tools with `mcp_servers_list_tools`
3. Assign them to the agent with `agent_tools_update`

### 6. Chat with the agent

1. Create a session: `playground_sessions_create` with `agent_id` and a `title`
2. Send messages: `playground_chat` with `agent_id`, `session_id`, and `message`
3. View history: `playground_messages_list` with `agent_id` and `session_id`

---

## Copyable Prompt for Any LLM

If your LLM is connected to GenKitKraft's MCP server but your client doesn't support MCP prompts, copy and paste the following into your conversation:

<details>
<summary>Click to expand the full prompt</summary>

```
You are connected to a GenKitKraft MCP server. GenKitKraft is a self-hostable platform for configuring and running LLM agents. Through these MCP tools, you can create fully configured AI agents, assign them tools, and chat with them — all programmatically.

Available tools by category:

PROVIDERS: provider_types_list, providers_create, providers_list, providers_get, providers_update, providers_delete, providers_test
PROMPTS: prompts_create, prompts_list, prompts_get, prompts_update, prompts_delete
AGENTS: agents_create, agents_list, agents_get, agents_update, agents_delete
AGENT TOOLS: agent_tools_get, agent_tools_update
HTTP TOOLS: http_tools_create, http_tools_list, http_tools_get, http_tools_update, http_tools_delete
MCP SERVERS: mcp_servers_create, mcp_servers_list, mcp_servers_get, mcp_servers_update, mcp_servers_delete, mcp_servers_list_tools
PLAYGROUND: playground_sessions_create, playground_sessions_list, playground_sessions_delete, playground_messages_list, playground_chat
HEALTH: health_liveness, health_readiness

To create a new agent, follow this order:
1. provider_types_list — see available provider types
2. providers_create — register an LLM provider (name, provider_type, api_key). Save the returned id.
3. providers_test — verify the provider works
4. prompts_create — create a system prompt (name, content). Save the returned id. (optional)
5. agents_create — create the agent (name, provider_id, model_id, system_prompt_id). Save the returned id.
6. http_tools_create / mcp_servers_create — create tools for the agent (optional)
7. agent_tools_update — assign tools to the agent (optional)
8. playground_sessions_create — start a chat session (agent_id, title). Save the returned id.
9. playground_chat — send a message (agent_id, session_id, message)

Key notes:
- All IDs are UUIDs returned by create operations. Always use the exact ID from the response.
- Model IDs depend on the provider: gpt-4o (OpenAI), gemini-2.0-flash (Google AI), claude-sonnet-4-20250514 (Anthropic).
- agent_tools_update replaces the entire tool configuration — include all tools you want assigned.
- List operations accept limit (default 20, max 100) and offset for pagination.
```

</details>

## Complete Tool Reference

See the [MCP Quickstart](/docs/guides/mcp-quickstart#available-tools) for a full table of all available tools with descriptions.
