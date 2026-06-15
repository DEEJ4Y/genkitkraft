# GenKitKraft

Self-hostable platform for configuring and running LLM agents. Built on [Google Genkit](https://genkit.dev/docs/go/overview) (Go SDK). Configure providers, create agents with custom instructions, connect MCP tool servers, and expose everything through an OpenAI-compatible API.

## Features

- [x] **Any LLM provider**: Google AI, OpenAI, Anthropic, Vertex AI, Bedrock, Azure, xAI, DeepSeek, Ollama
- [x] **Agent builder UI**: Create agents with system prompts, model selection, temperature, tool config, and tool call limits
- [x] **Built-in tools**: `web_fetch` fetches any URL and returns it as Markdown — no external setup required
- [x] **MCP tool support**: Connect tools via MCP servers (SSE and Streamable HTTP transports)
- [x] **MCP server**: All management APIs exposed as [MCP tools](https://DEEJ4Y.github.io/genkitkraft/docs/guides/mcp-quickstart) — build agents on top of GenKitKraft with any MCP client
- [x] **OpenAI-compatible API**: `/v1/chat/completions` with streaming support, works with any OpenAI client
- [x] **Single binary**: Frontend embedded in the Go binary; SQLite by default (zero external deps), PostgreSQL, MySQL, and MariaDB supported for multi-instance deployments
- [ ] **Smart tool selection**: Three tool modes (manual, auto-search, hybrid) to avoid context pollution

## MCP Server — Build Agents Programmatically

GenKitKraft exposes all management APIs as MCP tools. Connect any MCP client (Claude Desktop, Cursor, custom agents) to create providers, agents, prompts, and tools — no UI needed.

```jsonc
// Claude Desktop config (~/.config/Claude/claude_desktop_config.json)
{
  "mcpServers": {
    "genkitkraft": {
      "command": "npx",
      "args": ["mcp-remote", "http://localhost:8080/mcp"]
    }
  }
}
```

A built-in **`create-agent`** prompt guides the LLM through the full workflow. See the [MCP Quickstart](https://DEEJ4Y.github.io/genkitkraft/docs/guides/mcp-quickstart) and [Agent Creation Guide](https://DEEJ4Y.github.io/genkitkraft/docs/guides/mcp-agent-creation-guide) for details.

## Docs

**[https://DEEJ4Y.github.io/genkitkraft/](https://DEEJ4Y.github.io/genkitkraft/)**

Developer docs live under `docs/`:

- [Hexagonal Architecture Guide](docs/hexagonal-architecture/README.md) - project structure, patterns, dependency rules
- [TypeSpec Guide](docs/api-spec/01-typespec-guide.md) - API contract definitions

API spec implementations and generated OpenAPI output are in `spec/`.

## License

MIT
