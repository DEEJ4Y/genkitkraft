# GenKitKraft Agent Creation Guide

You are connected to a GenKitKraft MCP server. GenKitKraft is a self-hostable platform for configuring and running LLM agents. Through these MCP tools, you can create fully configured AI agents, assign them tools, and chat with them — all programmatically.

## Available Tools

### Provider Management
| Tool | Description |
|------|-------------|
| `provider_types_list` | List supported LLM provider types (openai, google_ai, anthropic, etc.) and their requirements |
| `providers_create` | Register an LLM provider with API key and configuration |
| `providers_list` | List all configured providers |
| `providers_get` | Get details of a specific provider |
| `providers_update` | Update a provider's config (name, API key, base URL, enabled status) |
| `providers_delete` | Delete a provider |
| `providers_test` | Test connectivity to a provider |

### System Prompts
| Tool | Description |
|------|-------------|
| `prompts_create` | Create a system prompt (name + content) |
| `prompts_list` | List all prompts |
| `prompts_get` | Get a prompt by ID |
| `prompts_update` | Update a prompt's name or content |
| `prompts_delete` | Delete a prompt |

### Agent Management
| Tool | Description |
|------|-------------|
| `agents_create` | Create an agent (requires provider_id, model_id; optionally system_prompt_id and sampling params) |
| `agents_list` | List all agents |
| `agents_get` | Get agent details |
| `agents_update` | Update agent configuration |
| `agents_delete` | Delete an agent |

### Agent Tool Configuration
| Tool | Description |
|------|-------------|
| `agent_tools_get` | Get which tools (HTTP tools and MCP servers) are assigned to an agent |
| `agent_tools_update` | Assign HTTP tools and MCP server tools to an agent (replaces entire config) |

### HTTP Tools (custom API tools for agents)
| Tool | Description |
|------|-------------|
| `http_tools_create` | Create an HTTP tool (name, description, method, URL, headers, body template, input schema) |
| `http_tools_list` | List all HTTP tools |
| `http_tools_get` | Get an HTTP tool by ID |
| `http_tools_update` | Update an HTTP tool |
| `http_tools_delete` | Delete an HTTP tool |

### MCP Servers (external MCP servers for agents)
| Tool | Description |
|------|-------------|
| `mcp_servers_create` | Register an external MCP server (name, transport, URL, headers) |
| `mcp_servers_list` | List all registered MCP servers |
| `mcp_servers_get` | Get an MCP server by ID |
| `mcp_servers_update` | Update an MCP server |
| `mcp_servers_delete` | Delete an MCP server |
| `mcp_servers_list_tools` | List tools exposed by an MCP server (discover what tools it provides) |

### Playground (chat with agents)
| Tool | Description |
|------|-------------|
| `playground_sessions_create` | Create a new chat session for an agent |
| `playground_sessions_list` | List chat sessions for an agent |
| `playground_sessions_delete` | Delete a chat session |
| `playground_messages_list` | List messages in a chat session |
| `playground_chat` | Send a message to an agent and get a response |

### Health
| Tool | Description |
|------|-------------|
| `health_liveness` | Check if the server is alive |
| `health_readiness` | Check if the server is ready to serve requests |

---

## Step-by-Step: Create an Agent

Follow these steps in order. Each step depends on outputs from the previous one.

### Step 1: Check available provider types

```
Tool: provider_types_list
Input: {}
```

This returns supported provider types (e.g., `openai`, `google_ai`, `anthropic`) with their requirements (API key, base URL, etc.). Use this to know what `provider_type` to use in the next step.

### Step 2: Create a provider

```
Tool: providers_create
Input: {
  "name": "My OpenAI Provider",
  "provider_type": "openai",
  "api_key": "<user's API key>"
}
```

**Save the returned `id`** — you'll need it as `provider_id` when creating the agent.

For providers with custom endpoints (e.g., Azure OpenAI, local models), also set `base_url`.

You can verify connectivity with:
```
Tool: providers_test
Input: { "id": "<provider_id>" }
```

### Step 3: Create a system prompt (optional but recommended)

```
Tool: prompts_create
Input: {
  "name": "Customer Support Agent",
  "content": "You are a helpful customer support agent. Be polite, concise, and accurate. If you don't know the answer, say so honestly."
}
```

**Save the returned `id`** — you'll need it as `system_prompt_id` when creating the agent.

### Step 4: Create the agent

```
Tool: agents_create
Input: {
  "name": "Support Bot",
  "provider_id": "<provider_id from step 2>",
  "model_id": "gpt-4o",
  "system_prompt_id": "<prompt_id from step 3>"
}
```

**Save the returned `id`** — this is your agent ID.

Optional parameters:
- `temperature_enabled`: true/false — enable temperature sampling
- `temperature`: 0.0 to 2.0 — controls randomness
- `top_p_enabled`: true/false — enable top-p (nucleus) sampling
- `top_p`: 0.0 to 1.0 — cumulative probability threshold
- `top_k_enabled`: true/false — enable top-k sampling
- `top_k`: integer — number of top tokens to consider

### Step 5: Assign tools to the agent (optional)

If you want the agent to use external tools (HTTP APIs or MCP servers), configure them first:

**Create HTTP tools:**
```
Tool: http_tools_create
Input: {
  "name": "Weather API",
  "description": "Get current weather for a city",
  "method": "GET",
  "url": "https://api.weather.com/current?city={{city}}",
  "headers": [{"name": "Authorization", "value": "Bearer <key>"}],
  "input_schema": "{\"type\":\"object\",\"properties\":{\"city\":{\"type\":\"string\"}}}"
}
```

**Register external MCP servers:**
```
Tool: mcp_servers_create
Input: {
  "name": "My MCP Server",
  "transport": "streamable_http",
  "url": "http://localhost:8080/mcp"
}
```

Then discover MCP server tools:
```
Tool: mcp_servers_list_tools
Input: { "id": "<mcp_server_id>" }
```

**Assign tools to the agent:**
```
Tool: agent_tools_update
Input: {
  "agent_id": "<agent_id from step 4>",
  "http_tool_ids": ["<http_tool_id>"],
  "mcp_servers": [
    {
      "mcp_server_id": "<mcp_server_id>",
      "tool_names": ["tool_name_1", "tool_name_2"]
    }
  ]
}
```

### Step 6: Chat with the agent

**Create a chat session:**
```
Tool: playground_sessions_create
Input: {
  "agent_id": "<agent_id>",
  "title": "Test Session"
}
```

**Send a message:**
```
Tool: playground_chat
Input: {
  "agent_id": "<agent_id>",
  "session_id": "<session_id>",
  "message": "Hello! Can you help me?"
}
```

The response includes the agent's reply. Continue the conversation by sending more messages to the same session.

**View conversation history:**
```
Tool: playground_messages_list
Input: {
  "agent_id": "<agent_id>",
  "session_id": "<session_id>"
}
```

---

## Quick Reference: Required ID Chain

```
provider_types_list → pick a type
     ↓
providers_create (type + API key) → provider_id
     ↓
prompts_create (name + content) → prompt_id (optional)
     ↓
agents_create (provider_id + model_id + prompt_id) → agent_id
     ↓
http_tools_create / mcp_servers_create → tool IDs (optional)
     ↓
agent_tools_update (agent_id + tool IDs) (optional)
     ↓
playground_sessions_create (agent_id) → session_id
     ↓
playground_chat (agent_id + session_id + message) → agent response
```

## Tips

- **IDs are UUIDs**: All created resources return a UUID `id`. Always use the exact ID from the creation response.
- **Pagination**: List operations accept `limit` and `offset` parameters. Defaults: limit=20, max=100.
- **Updating vs replacing**: `agents_update` and similar tools do partial updates — only fields you include are changed. `agent_tools_update` is an exception — it replaces the entire tool configuration.
- **Model IDs**: The `model_id` depends on the provider type. Common examples: `gpt-4o`, `gpt-4o-mini` (OpenAI), `gemini-2.0-flash` (Google AI), `claude-sonnet-4-20250514` (Anthropic).
- **Testing providers**: Always test a provider after creation to verify the API key and configuration are correct.
- **Multiple agents**: You can reuse the same provider and prompt across multiple agents. Create once, reference by ID.
