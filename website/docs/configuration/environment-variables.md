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
| `CACHE_PROVIDER` | Cache backend: `memory`, `redis`, `valkey` | `memory` | No |
| `CACHE_URL` | Connection URL for non-memory providers | — | When `CACHE_PROVIDER` ≠ `memory` |
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

### `CACHE_PROVIDER`

Selects the backend for cached state. Defaults to `memory`.

| Value | Backend | Use case |
|---|---|---|
| `memory` | In-process (default) | Single-node, zero-config |
| `redis` | Redis 6+ | Multi-instance |
| `valkey` | Valkey 7+ | Multi-instance |

The cache holds two kinds of state:

- **Session tokens** (24h TTL) — issued on login and checked on every authenticated request.
- **Login rate-limit counters** (1 minute window) — at most 5 failed attempts per IP.

Web-fetch results are **not** kept here. The built-in web-fetch tool caches into a process-local store regardless of `CACHE_PROVIDER`, so agent tool traffic can never evict session tokens or consume the shared cache's memory. Each instance keeps its own web-fetch cache; a repeated fetch on a cold instance simply fetches again.

With `memory`, both are process-local. That is correct for a single instance, but **running more than one instance on `memory` breaks authentication**: a login served by instance A is unknown to instance B, so the next request returns `401`. Rate limiting degrades the same way — each instance counts failures separately, so N instances allow roughly N times the intended attempts.

Set `CACHE_PROVIDER` to `redis` or `valkey` for any multi-instance deployment. Both speak the same protocol and are handled identically; the two values exist only to describe your infrastructure.

An unrecognised value is rejected at startup rather than falling back to `memory`, since a silent fallback would reintroduce exactly the problems above.

:::note
The cache is not durable storage. Losing it logs users out and clears rate-limit counters, but no application data is affected — that lives in the database. Persistence is not required.

**Use `noeviction`** (the Redis/Valkey default) and size `maxmemory` for your expected number of concurrent sessions.

`volatile-ttl` offers session tokens no protection here: every key GenKitKraft writes has a TTL, so the "only evict volatile keys" carve-out excludes nothing and the policy evicts from the whole keyspace. It is in fact the worse choice — `volatile-ttl` evicts the shortest remaining TTL first, and the shortest TTLs are the 1-minute rate-limit counters. Under memory pressure it quietly resets brute-force protection before it starts dropping sessions, and then drops sessions too.

With `noeviction`, reads keep working when the cache is full, so existing sessions stay valid; writes fail, so new logins return `503` until memory frees up. That is a visible, honest failure rather than a silent one.

Sizing: a session is roughly 200 bytes including overhead, so even `maxmemory 64mb` holds far more concurrent sessions than a self-hosted instance is likely to see.
:::

### `CACHE_URL`

Connection URL for non-memory providers. Required when `CACHE_PROVIDER` is anything other than `memory` — the server refuses to start without it. The connection is verified at startup, so a wrong URL fails immediately rather than at the first request.

**Redis**

```bash
CACHE_URL=redis://host:6379
```

With a password, a database index, or TLS:

```bash
CACHE_URL=redis://:password@host:6379/0
CACHE_URL=rediss://:password@host:6379/0   # TLS
```

**Valkey**

Valkey accepts the same `redis://` URLs. The `valkey://` and `valkeys://` schemes are also accepted and treated as equivalent to `redis://` and `rediss://`:

```bash
CACHE_URL=valkey://host:6379
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
