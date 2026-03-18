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

Everything lives under `docs/`:

- [Hexagonal Architecture Guide](docs/hexagonal-architecture/README.md) - project structure, patterns, dependency rules
- [TypeSpec Guide](docs/api-spec/01-typespec-guide.md) - API contract definitions

API spec implementations and generated OpenAPI output are in `spec/`.

## Project Structure

```
/
├── cmd/server/main.go            # entrypoint, embeds ui/dist
├── internal/
│   ├── domain/                   # entities and business logic
│   │   ├── agent.go              # agent entity (prompt, model, tool config)
│   │   ├── provider.go           # LLM provider entity
│   │   └── mcp.go                # MCP connection entity
│   ├── ports/                    # interfaces (inbound + outbound)
│   │   ├── api.go                # OpenAI-compat API port
│   │   ├── admin.go              # admin REST API port
│   │   ├── generation.go         # LLM generation port
│   │   └── repository.go         # persistence port
│   └── adapters/
│       ├── http/                 # OpenAI-compat API + admin API handlers
│       ├── genkit/               # Genkit SDK adapter (generation, MCP, tool search)
│       └── persistence/          # SQLite/Postgres repository implementations
├── ui/                           # frontend SPA (admin config UI)
│   ├── src/
│   ├── dist/                     # build output, embedded into Go binary
│   └── package.json
├── docs/                         # architecture and API spec docs
├── spec/                         # TypeSpec definitions and generated OpenAPI
├── Makefile
├── Dockerfile
├── docker-compose.yml
└── docker-compose.postgres.yml
```

## Base Spec

Health check server with two endpoints:

- `GET /readyz` - readiness probe
- `GET /livez` - liveness probe

## License

MIT
