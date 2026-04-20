---
sidebar_position: 1
---

# Environment Variables

GenKitKraft is configured entirely through environment variables. No config files needed.

## Reference Table

| Variable | Description | Default | Required |
|---|---|---|---|
| `PORT` | HTTP server port | `8080` | No |
| `DATABASE_PATH` | Path to SQLite database file | `/data/app.db` | No |
| `ENCRYPTION_KEY` | Secret key for encrypting provider API keys at rest | — | **Yes** |
| `AUTH_CREDENTIALS` | Comma-separated `username:password` pairs | _(unset — auth disabled)_ | No |
| `PUBLIC_API_KEY` | Comma-separated API keys for deploy endpoints | _(unset — deploy is public)_ | No |

### `PORT`

The port the HTTP server listens on. Both the API and the embedded UI are served on this port.

### `DATABASE_PATH`

File path for the SQLite database. When running in Docker, ensure this path is on a persistent volume (`/data` by default) to prevent data loss on container recreation.

### `ENCRYPTION_KEY`

A secret string used to encrypt LLM provider API keys before storing them in the database. **The server will refuse to start if this variable is not set.**

Tips:

- Generate a strong key: `openssl rand -base64 32`
- Keep it safe — if lost, existing provider configurations become unreadable
- Changing this key requires re-creating all provider configurations

### `AUTH_CREDENTIALS`

Controls login-based authentication. Format: `username:password` pairs separated by commas.

```bash
AUTH_CREDENTIALS=admin:changeme,readonly:viewer123
```

- When set: all UI and API access requires authentication
- When unset: authentication is disabled entirely
- See [Authentication](./authentication) for more details

### `PUBLIC_API_KEY`

Controls API key authentication for the [Deploy API](/docs/api/deploy) endpoints (both stateless and stateful sessions). Format: one or more keys separated by commas.

```bash
PUBLIC_API_KEY=sk-my-secret-key
# Or multiple keys:
PUBLIC_API_KEY=sk-key-one,sk-key-two
```

- When set: deploy requests must include `Authorization: Bearer <key>`
- When unset: all deploy endpoints are publicly accessible (no authentication required)
- This is separate from `AUTH_CREDENTIALS` — deploy API keys are for external integrations, while auth credentials protect the management UI
