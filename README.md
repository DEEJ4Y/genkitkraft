# GenKitKraft

Self-hostable platform for configuring and running LLM agents. Built on [Google Genkit](https://genkit.dev/docs/go/overview) (Go SDK). Configure providers, create agents with custom instructions, connect MCP tool servers, and expose everything through an OpenAI-compatible API.

## Features

- [ ] **Any LLM provider**: Google AI, OpenAI, Anthropic, Vertex AI, Bedrock, Azure, xAI, DeepSeek, Ollama
- [ ] **Agent builder UI**: Create agents with system prompts, model selection, temperature, and tool config
- [ ] **MCP tool support**: Connect tools via MCP servers (stdio and SSE transports)
- [ ] **Smart tool selection**: Three tool modes (manual, auto-search, hybrid) to avoid context pollution
- [ ] **OpenAI-compatible API**: `/v1/chat/completions` with streaming support, works with any OpenAI client
- [ ] **Single binary**: Frontend embedded in the Go binary, SQLite by default, zero external dependencies

## Docs

**[https://DEEJ4Y.github.io/genkitkraft/](https://DEEJ4Y.github.io/genkitkraft/)**

Developer docs live under `docs/`:

- [Hexagonal Architecture Guide](docs/hexagonal-architecture/README.md) - project structure, patterns, dependency rules
- [TypeSpec Guide](docs/api-spec/01-typespec-guide.md) - API contract definitions

API spec implementations and generated OpenAPI output are in `spec/`.

## License

MIT
