---
sidebar_position: 3
---

# MCP Quickstart

Get GenKitKraft connected to your MCP client in under two minutes. Once connected, you can create agents, configure providers, and chat — all from your AI assistant.

## Prerequisites

- GenKitKraft running (see [Installation](./installation))

## 1. Add GenKitKraft to your MCP client

The MCP server is available at `http://localhost:8080/mcp` using the **Streamable HTTP** transport.

### Claude Desktop

Add this to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "genkitkraft": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Cursor

In Cursor settings, add an MCP server:

- **Type**: Streamable HTTP
- **URL**: `http://localhost:8080/mcp`

### Other MCP clients

Any client that supports Streamable HTTP transport can connect using the URL above.

:::tip Authentication
If you've set [`AUTH_CREDENTIALS`](/docs/configuration/environment-variables#auth_credentials), add a Basic Auth header:

```json
"headers": {
  "Authorization": "Basic <base64-encoded username:password>"
}
```

Generate the value with: `echo -n "admin:changeme" | base64`
:::

## 2. Create your first agent

Once connected, ask your AI assistant:

> "Set up a new agent using OpenAI GPT-4o with a helpful assistant prompt."

Your MCP client will call GenKitKraft's tools automatically to:

1. **Create a provider** — register your OpenAI API key
2. **Create a prompt** — set up a system prompt
3. **Create the agent** — wire everything together
4. **Start chatting** — open a playground session

The built-in **`create-agent` prompt** guides supported clients (like Claude Desktop) through the full workflow automatically.

## What's available

GenKitKraft exposes **40+ MCP tools** covering:

| Category | What you can do |
|----------|----------------|
| **Agents** | Create, update, delete, and list agents |
| **Providers** | Configure LLM providers (OpenAI, Anthropic, Google, etc.) |
| **Prompts** | Manage system prompts |
| **Tools** | Add HTTP tools and MCP servers, assign them to agents |
| **Playground** | Chat with agents, manage sessions |
| **Health** | Check server status |

## Next steps

- [Full MCP reference](/docs/guides/mcp-quickstart) — complete tool list, auth details, and troubleshooting
- [Agent creation guide](/docs/guides/mcp-agent-creation-guide) — step-by-step walkthrough with input schemas
- [Tools guide](/docs/guides/tools) — configure HTTP tools and external MCP servers for your agents
