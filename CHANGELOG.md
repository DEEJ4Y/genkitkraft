# Changelog

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
