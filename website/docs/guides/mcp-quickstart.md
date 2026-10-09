---
sidebar_position: 6
---

# MCP Quickstart

GenKitKraft exposes all of its management APIs as [MCP](https://modelcontextprotocol.io/) tools. This lets you manage agents, providers, prompts, tools, and even chat — directly from any MCP-compatible client (Claude Desktop, Cursor, custom agents, etc.).

## Prerequisites

- GenKitKraft running (see [Installation](/docs/getting-started/installation))

## Endpoint

The MCP server is available at:

```
http://<host>:<port>/mcp
```

Default: `http://localhost:8080/mcp`

It uses the **Streamable HTTP** transport.

## Authentication

- If [`AUTH_CREDENTIALS`](/docs/configuration/environment-variables#auth_credentials) is set, the MCP endpoint requires **HTTP Basic Auth** with the same username and password.
- If `AUTH_CREDENTIALS` is not set, the MCP endpoint is open (no auth).

## Connecting from Claude Desktop

Claude Desktop connects to a remote server through the `mcp-remote` bridge. This needs [Node.js](https://nodejs.org) (for `npx`). Add GenKitKraft to your Claude Desktop MCP config (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "genkitkraft": {
      "command": "npx",
      "args": ["mcp-remote", "http://localhost:8080/mcp"]
    }
  }
}
```

If authentication is enabled, pass an `Authorization` header to `mcp-remote`:

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
        "AUTH_HEADER": "Basic <base64-encoded username:password>"
      }
    }
  }
}
```

:::tip
Generate the base64 value with: `echo -n "admin:changeme" | base64`
:::

## Connecting from Cursor

In Cursor settings, add an MCP server with:

- **Type**: Streamable HTTP
- **URL**: `http://localhost:8080/mcp`

If auth is required, add this header:

```json
"headers": {
  "Authorization": "Basic <base64-encoded username:password>"
}
```

## Available Tools

Once connected, your MCP client will discover all available tools automatically. Here's a summary:

### Agents

| Tool | Description |
|------|-------------|
| `agents_list` | List all agents with pagination |
| `agents_get` | Get an agent by ID |
| `agents_create` | Create a new agent |
| `agents_update` | Update an agent |
| `agents_delete` | Delete an agent |

### Agent Tool Config

| Tool | Description |
|------|-------------|
| `agent_tools_get` | Get which tools are assigned to an agent |
| `agent_tools_update` | Update tool assignments for an agent |

### Prompts

| Tool | Description |
|------|-------------|
| `prompts_list` | List all system prompts |
| `prompts_get` | Get a prompt by ID |
| `prompts_create` | Create a new system prompt |
| `prompts_update` | Update a prompt |
| `prompts_delete` | Delete a prompt |

### Providers

| Tool | Description |
|------|-------------|
| `providers_list` | List all LLM providers |
| `providers_get` | Get a provider by ID |
| `providers_create` | Create a new provider |
| `providers_update` | Update a provider |
| `providers_delete` | Delete a provider |
| `providers_test` | Test provider connectivity |
| `provider_types_list` | List supported provider types |

### HTTP Tools

| Tool | Description |
|------|-------------|
| `http_tools_list` | List all HTTP tools |
| `http_tools_get` | Get an HTTP tool by ID |
| `http_tools_create` | Create a new HTTP tool |
| `http_tools_update` | Update an HTTP tool |
| `http_tools_delete` | Delete an HTTP tool |

### MCP Servers

| Tool | Description |
|------|-------------|
| `mcp_servers_list` | List all registered MCP servers |
| `mcp_servers_get` | Get an MCP server by ID |
| `mcp_servers_create` | Register a new MCP server |
| `mcp_servers_update` | Update an MCP server |
| `mcp_servers_delete` | Delete an MCP server |
| `mcp_servers_list_tools` | Discover tools from an MCP server |

### Playground

| Tool | Description |
|------|-------------|
| `playground_sessions_list` | List chat sessions for an agent |
| `playground_sessions_create` | Create a new chat session (`title` is optional) |
| `playground_sessions_delete` | Delete a chat session |
| `playground_messages_list` | List messages in a session |
| `playground_chat` | Send a message and get a response |

`playground_chat` and `playground_messages_list` need a session ID. Send it as `session_id`. If your client cannot send `session_id`, send the same value as `fallback_session_id`. When both fields are set, `fallback_session_id` takes precedence.

`playground_chat` accepts the same optional overrides as the HTTP API: provider, model, system prompt, sampling settings (`temperature`, `top_p`, `top_k`), `max_tool_calls`, `http_tool_ids`, `mcp_servers` and `built_in_tool_ids`. An override applies to that request only. It does not change the saved agent.

### Gaps

| Tool | Description |
|------|-------------|
| `gaps_list` | List the gaps an agent reported, with pagination |
| `gaps_get` | Get a gap by ID |
| `gaps_resolve` | Mark a gap as resolved |
| `gaps_dismiss` | Dismiss a gap with a reason category. A gap dismissed as `unrelated` can never be reopened |
| `gaps_reopen` | Reopen a resolved or dismissed gap |

See the [Gaps guide](/docs/guides/gaps) for more about gaps.

### Built-in Tools

| Tool | Description |
|------|-------------|
| `built_in_tools_list` | List all available built-in tools. Use the IDs with `agent_tools_update` (`built_in_tool_ids`) |

### Auth

| Tool | Description |
|------|-------------|
| `auth_login` | Log in with username/password |
| `auth_logout` | Log out (invalidate session) |
| `auth_get_me` | Get current user info |
| `auth_get_status` | Check if auth is required |

### Health

| Tool | Description |
|------|-------------|
| `health_liveness` | Liveness check |
| `health_readiness` | Readiness check |

## Example: Set up a new agent via MCP

Using any MCP client, you can create a fully configured agent in a few tool calls:

1. **`provider_types_list`** — see available provider types
2. **`providers_create`** — configure an LLM provider (e.g., OpenAI with your API key)
3. **`prompts_create`** — create a system prompt
4. **`agents_create`** — create an agent using the provider and prompt
5. **`playground_sessions_create`** — start a chat session
6. **`playground_chat`** — chat with your agent

For a detailed walkthrough with examples and input schemas, see the [Agent Creation Guide](/docs/guides/mcp-agent-creation-guide).

## Built-in Prompts

The MCP server includes these prompts. MCP clients that support server-side prompts (like Claude Desktop) can load them automatically.

| Prompt | Description |
|--------|-------------|
| `create-agent` | A guide to all available tools and the correct workflow to create agents |
| `backup` | A guide to back up all configurations (tools, MCP servers, prompts, agents) to a structured markdown file |
| `restore` | A guide to restore configurations from a backup file, with conflict detection and resolution |

## Troubleshooting

**401 Unauthorized**: `AUTH_CREDENTIALS` is set but your client isn't sending the correct Basic Auth header. Double-check the base64-encoded `username:password`.

**Connection refused**: Make sure GenKitKraft is running and the URL/port are correct. If running in Docker, ensure the port is mapped.

**A tool says `session_id` is missing, but the client sent it**: Some clients drop the `session_id` argument before it reaches the server. Send the same value as `fallback_session_id`.

**Empty tool list**: The MCP client may not support the Streamable HTTP transport. Check your client's MCP transport support.
