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

| Variable           | Description                                          | Default                   | Required |
| ------------------ | ---------------------------------------------------- | ------------------------- | -------- |
| `PORT`             | HTTP server port                                     | `8080`                    | No       |
| `DATABASE_PATH`    | Path to the SQLite database file                     | `/data/app.db`            | No       |
| `ENCRYPTION_KEY`   | Secret key used to encrypt provider API keys at rest | —                         | **Yes**  |
| `AUTH_CREDENTIALS` | Comma-separated `user:pass` pairs                    | _(unset — auth disabled)_ | No       |

### `PORT`

The port the HTTP server listens on. Both the API and the UI are served on this port.

### `DATABASE_PATH`

The file path where GenKitKraft stores its SQLite database. Make sure this path is on a persistent volume when running in Docker, otherwise data will be lost when the container is recreated.

### `ENCRYPTION_KEY`

A secret string used to encrypt LLM provider API keys before they are stored in the database. This ensures API keys are protected at rest — even if the database file is compromised, the keys cannot be read without this value.

**The server will refuse to start if this variable is not set.**

Choose a long, random string (e.g. generated with `openssl rand -base64 32`) and keep it safe. If you lose or change this key, existing provider configurations will become unreadable and you will need to re-create them.

```bash
ENCRYPTION_KEY=my-random-secret-key-here
```

### `AUTH_CREDENTIALS`

Controls login-based authentication for the UI and API. The value is a comma-separated list of `username:password` pairs:

```bash
AUTH_CREDENTIALS=admin:changeme,user2:password2
```

When set, all API and UI access requires logging in. When unset, authentication is disabled entirely.
