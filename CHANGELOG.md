# Changelog

## v0.6.1 — MCP discovery timeout and horizontal scaling docs

### Fixes

- **MCP server discovery no longer hangs on an unresponsive server** — connecting to a user-configured MCP server had no timeout: the underlying genkit MCP client always dials with `context.Background()` internally, so a transport-level HTTP timeout was the only way to bound the call. Both the SSE and Streamable HTTP transports now share the same 30s-timeout client already used by `web_fetch` and the custom HTTP tool. Covered by a new integration test (`internal/adapters/mcp_discovery/timeout_test.go`).

### Docs

- **New Horizontal Scaling guide** — a single page tying together the three requirements for running multiple instances (shared database, shared cache, and an identical `ENCRYPTION_KEY` on every instance — the last of which wasn't documented anywhere before) and what already works with zero extra config (migration locking, cross-instance cancel, `Last-Event-ID` SSE resume).

## v0.6.0 — Cache-backed cross-instance stream cancellation and migration locking

### Fixes

- **Playground "stop generation" now reaches other instances** — the SSE stream-cancellation registry was process-local only, so a "stop" request landing on a different instance than the one running the generation silently no-op'd, with no config to fix it. It's now backed by the same pluggable cache port used by sessions and login rate limiting (`internal/adapters/cache_stream_registry`): the actual cancel still runs locally (a `context.CancelFunc` can't cross a process boundary), but a short-TTL signal relayed through the shared cache lets the owning instance notice a remote cancel request. `CACHE_PROVIDER=redis`/`valkey` now makes cross-instance "stop" work; `CACHE_PROVIDER=memory` (the default) degrades to the previous process-local behavior for free.

- **Migration locking** — Multiple instances starting simultaneously against a PostgreSQL, MySQL, or MariaDB database with pending migrations no longer race and crash. `postgres_db` now uses goose's built-in session-level advisory lock; `mysql_db`/`mariadb` use a new `GET_LOCK()`/`RELEASE_LOCK()`-based session locker. Losing instances block until the migrating instance finishes, then proceed normally — none exit. SQLite is unaffected (single-node by design).

## v0.5.0 — Pluggable database adapters

### New features

- **Pluggable database adapters** — GenKitKraft can now run against SQLite (default), PostgreSQL, MySQL, or MariaDB with no application-level code changes. Select the engine via the `DATABASE_PROVIDER` environment variable; provide a DSN via `DATABASE_URL` for non-SQLite providers.
  - `sqlite` — file-based, zero-config, default for local development
  - `postgres` — recommended for production and multi-instance deployments
  - `mysql` / `mariadb` — full feature parity with PostgreSQL adapter

- **Database migrations on startup** — Goose migrations run automatically when the server starts. If a migration fails, the server refuses to start.

- **Connection pool tuning** — MySQL and PostgreSQL connections now set `ConnMaxLifetime(5m)` alongside `MaxOpenConns` and `MaxIdleConns`, preventing silent failures caused by connections outliving the server's `wait_timeout`.

### Testing

- **Integration tests** — Every adapter has integration tests using Testcontainers. Tests run against real database containers and cover CRUD operations, not-found errors, foreign key constraints, and unique-constraint handling.

- **CI pipeline** — Three GitHub Actions jobs run on every PR:
  - **Build** — `go build ./cmd/server/...`
  - **Vet** — `go vet ./...`
  - **Test** — `go test -tags integration -v ./...` (includes Testcontainers integration tests)

To run integration tests locally: `make test-integration` (requires Docker).

### Dependencies added

| Package | Version |
|---|---|
| `github.com/go-sql-driver/mysql` | v1.10.0 |
| `github.com/jackc/pgx/v5` | v5.10.0 |
| `github.com/testcontainers/testcontainers-go` | v0.42.0 |
| `github.com/stretchr/testify` | v1.11.1 |

---

## v0.3.2 — Built-in tools and tool call limits

### New features

- **`web_fetch` built-in tool** — Agents can now fetch any URL and receive its content as Markdown. Fetches static HTML; does not execute JavaScript or render dynamic content. Enable it per-agent on the Tools tab or via `builtInToolIds: ["web_fetch"]` in the API.

- **Tool call limits (`MaxToolCalls`)** — Agents now support a configurable cap on the number of tool call iterations per request. Prevents runaway loops when tools trigger further tool calls. Default is 10; configurable in the agent's Generation Parameters.

- **URL sanitization** — All URLs used by tools are now percent-encoded before use, preventing errors caused by unsafe characters in paths or query strings.
