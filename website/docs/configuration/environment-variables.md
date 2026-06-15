---
sidebar_position: 1
---

# Environment Variables

GenKitKraft is configured entirely through environment variables. No config files needed.

## Reference Table

| Variable | Description | Default | Required |
|---|---|---|---|
| `PORT` | HTTP server port | `8080` | No |
| `DATABASE_PROVIDER` | Database engine: `sqlite`, `postgres`, `mysql`, `mariadb` | `sqlite` | No |
| `DATABASE_PATH` | Path to SQLite database file | `/data/app.db` | No |
| `DATABASE_URL` | Connection URL/DSN for non-SQLite providers | — | When `DATABASE_PROVIDER` ≠ `sqlite` |
| `ENCRYPTION_KEY` | Secret key for encrypting provider API keys at rest | — | **Yes** |
| `AUTH_CREDENTIALS` | Comma-separated `username:password` pairs | _(unset — auth disabled)_ | No |
| `PUBLIC_API_KEY` | Comma-separated API keys for deploy endpoints | _(unset — deploy is public)_ | No |

### `PORT`

The port the HTTP server listens on. Both the API and the embedded UI are served on this port.

### `DATABASE_PROVIDER`

Selects the database engine. Defaults to `sqlite`.

| Value | Engine | Use case |
|---|---|---|
| `sqlite` | SQLite (default) | Single-node, zero-config |
| `postgres` | PostgreSQL 14+ | Multi-instance, recommended for production |
| `mysql` | MySQL 8.0+ | Multi-instance |
| `mariadb` | MariaDB 10.6+ | Multi-instance |

Database migrations run automatically on startup. If a migration fails, the server will not start.

SQLite is write-serialised (`MaxOpenConns=1`) and unsuitable for multiple instances sharing storage. Switch to `postgres`, `mysql`, or `mariadb` for horizontal scaling.

### `DATABASE_URL`

Connection string for non-SQLite providers. Required when `DATABASE_PROVIDER` is anything other than `sqlite` — the server will refuse to start if it is missing.

**PostgreSQL**

```bash
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=require
```

For local or internal networks without TLS:

```bash
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=disable
```

**MySQL**

```bash
# parseTime=true is required for correct timestamp handling
DATABASE_URL=user:password@tcp(host:3306)/dbname?parseTime=true
```

**MariaDB**

```bash
DATABASE_URL=user:password@tcp(host:3306)/dbname?parseTime=true
```

Non-SQLite providers use a connection pool (25 max open, 5 max idle). Size this according to your database server's `max_connections` limit.

### `DATABASE_PATH`

File path for the SQLite database. Only used when `DATABASE_PROVIDER=sqlite` (the default).

When running in Docker, mount this path on a persistent volume to prevent data loss on container recreation:

```yaml
environment:
  - DATABASE_PATH=/data/app.db
volumes:
  - genkitkraft-data:/data
```

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
- Also controls authentication for the [MCP endpoint](/docs/guides/mcp-quickstart) — when set, the MCP server requires HTTP Basic Auth with the same credentials
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
