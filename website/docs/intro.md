---
sidebar_position: 1
slug: /
---

# Introduction

GenKitKraft is a self-hostable platform for configuring and running LLM agents, built on [Google Genkit](https://github.com/firebase/genkit) (Go SDK).

## Features

- **Agent Builder UI** — Create agents with custom system instructions
- **MCP Tool Support** — Connect Model Context Protocol tool servers to your agents
- **Multi-Provider LLM Access** — Use OpenAI, Anthropic, Google, and more
- **OpenAI-Compatible API** — Expose configured agents via a standard API
- **Single Binary Deployment** — Server and UI ship as one binary

## Quick Start

### Using Docker Compose

```bash
docker compose up --build
```

The server starts on port `8080` by default.

### From Source

```bash
# Build the UI
cd ui && npm ci && npm run build && cd ..

# Build and run the server
go build ./cmd/server/...
PORT=8080 ./server
```

## Configuration

GenKitKraft is configured via environment variables:

| Variable           | Description                       | Default                   |
| ------------------ | --------------------------------- | ------------------------- |
| `PORT`             | HTTP server port                  | `8080`                    |
| `AUTH_CREDENTIALS` | Comma-separated `user:pass` pairs | _(unset — auth disabled)_ |

### Authentication

To enable authentication, set the `AUTH_CREDENTIALS` environment variable:

```bash
AUTH_CREDENTIALS=admin:changeme,user2:password2
```

When set, all API and UI access requires logging in. When unset, authentication is disabled entirely.
