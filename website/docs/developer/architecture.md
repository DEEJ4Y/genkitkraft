---
sidebar_position: 1
---

# Architecture

GenKitKraft follows **Hexagonal Architecture** (Ports & Adapters), where business logic sits at the center, completely isolated from external concerns. The core communicates with the outside world only through well-defined interfaces (ports), and concrete implementations (adapters) plug into those interfaces. This makes the system testable, flexible, and easy to extend — you can swap databases, APIs, or UI frameworks without touching business logic.

> For deep dives on each topic, see the full guide in [`docs/hexagonal-architecture/`](https://github.com/DEEJ4Y/genkitkraft/tree/main/docs/hexagonal-architecture).

## Layer Overview

```
┌─────────────────────────────────────────────────┐
│                   Services                       │
│              (Composition Root)                  │
│  Wires everything together, dependency injection │
├──────────┬──────────────────────┬────────────────┤
│ Handlers │    Application       │   Adapters     │
│ (HTTP,   │  (Commands/Queries)  │ (DB, external  │
│  gRPC)   │   Use case logic     │  services)     │
├──────────┴──────────────────────┴────────────────┤
│                    Ports                          │
│          (Interface definitions)                  │
├──────────────────────────────────────────────────┤
│                   Domain                          │
│    (Entities, value objects, business rules)      │
│              Pure Go — no imports                 │
└──────────────────────────────────────────────────┘
```

| Layer | Purpose |
|---|---|
| **Domain** | Pure business entities, value objects, and rules. No external imports — stdlib only. |
| **Ports** | Interface definitions with port-specific DTOs. The contracts between layers. |
| **Application** | Use cases (commands for writes, queries for reads). Orchestrates ports. |
| **Adapters** | Concrete implementations of ports (database, external APIs, etc.). |
| **Handlers** | Primary adapters that translate HTTP/gRPC/CLI requests into application calls. |
| **Services** | Composition root — the only place that knows all layers. Wires dependencies. |

## Dependency Flow Rules

This is the most critical rule in the codebase. Violating it breaks the architecture.

| Package | Can Import | Cannot Import |
|---|---|---|
| `domain/` | standard library only | everything else |
| `ports/` | `domain/` | `adapters/`, `app/`, `handlers/` |
| `app/` | `ports/`, `domain/` | `adapters/`, `handlers/` |
| `adapters/` | `ports/`, `domain/`, `clients/` | `app/`, `handlers/` |
| `handlers/` | `app/`, `domain/`, `common/` | `adapters/` |
| `services/` | all internal packages | — |
| `cmd/` | `services/`, `config/` | — |

**Rule of thumb:** If you find yourself importing an adapter inside `app/`, or `app/` inside an adapter — stop. Define a port interface, implement it in an adapter, and inject through the composition root.

## Directory Structure

```
cmd/                        → Entry points (main.go)
internal/
  domain/                   → Pure business entities, value objects, rules
  ports/
    <port_name>/
      interface.go          → Interface definition
      types.go              → Port-specific param/result DTOs
  adapters/
    <adapter_name>/
      <impl>.go             → Implementation
      type_conversion.go    → Mapping between port DTOs and infra types
  app/
    commands/               → Write operations (one command per file)
    queries/                → Read operations (one query per file)
    decorators/             → Cross-cutting wrappers (logging, tracing, etc.)
    executors/              → Generic Executor interfaces
    <name>_app.go           → Application struct grouping commands & queries
  handlers/
    <handler_name>/
      <service>.go
      type_conversion.go
      interceptors/         → Middleware (auth, logging, correlation ID)
  clients/                  → Low-level infrastructure client wrappers
  common/                   → Shared utilities (errors, logger, metrics)
  config/                   → Configuration structs (from env vars)
  services/                 → Composition root (DI wiring)
resources/test/             → Test infrastructure (containers, seed, mocks)
```

## Key Patterns

### Executor Pattern

All use cases implement a generic interface, enabling uniform decorator wrapping:

```go
type Executor[Params any] interface {
    Execute(ctx context.Context, params Params) error
}

type ExecutorWithReturn[Params, Result any] interface {
    Execute(ctx context.Context, params Params) (Result, error)
}
```

Each command or query is a struct that implements one of these interfaces. This keeps use cases focused on a single responsibility.

### Application Struct

Commands and queries are grouped into an Application struct for convenient injection:

```go
type AdminApp struct {
    Commands AdminCommands
    Queries  AdminQueries
}

type AdminCommands struct {
    CreateUser commands.CreateUserHandler
}

type AdminQueries struct {
    GetUser queries.GetUserHandler
}
```

Handlers receive the Application struct — they never interact with ports directly.

### Decorator Pattern

Decorators wrap executors for cross-cutting concerns like logging, tracing, caching, and error handling. They are applied in the composition root, and order matters (outermost executes first):

```go
// In the composition root
createUser := decorators.WithLogging(
    decorators.WithErrorHandling(
        commands.NewCreateUser(userRepo),
    ),
    logger,
)
```

### Compile-Time Interface Checks

Every adapter must include a compile-time assertion to ensure it satisfies its port:

```go
var _ portpkg.SomeInterface = (*AdapterImpl)(nil)
```

This catches interface drift at compile time rather than runtime.

### Type Conversion

Each adapter and handler has its own `type_conversion.go` file. Infrastructure types (database models, API responses) are never leaked into ports or domain. Conversion happens at the boundary:

- **Adapters:** Convert between port DTOs and infrastructure types (e.g., SQL rows ↔ port structs)
- **Handlers:** Convert between HTTP request/response types and application DTOs

### Manual Dependency Injection

No DI frameworks. The composition root (`internal/services/`) is the **only** place that knows all layers. It constructs adapters, wraps them in decorators, builds application structs, and injects them into handlers.

### Error Handling

Use `AppError` with typed error codes from `internal/common/errors/`:

```go
common.NewAppError(common.NotFound, "user not found")
common.NewAppError(common.InvalidInput, "email is required")
common.NewAppError(common.Conflict, "username already taken")
```

Error codes: `NotFound`, `InvalidInput`, `Conflict`, `Unauthorized`, `Forbidden`, `Internal`.

Handlers map `AppError` codes to HTTP status codes. The error handler decorator wraps unexpected panics/errors as `Internal`.

### Configuration

All configuration comes from environment variables, loaded once at startup, and injected through the composition root. Adapters receive only the config values they need — never the full config struct.

## Testing

### Unit Tests

Mock port interfaces to test the application layer in isolation. No infrastructure needed:

```go
mockRepo := &mock.UserRepository{}
mockRepo.On("FindByID", ctx, "user-1").Return(expectedUser, nil)

handler := queries.NewGetUser(mockRepo)
result, err := handler.Execute(ctx, "user-1")
```

Mocks live in `resources/test/mock/` and must include compile-time interface checks.

### Integration Tests

Test adapters against real infrastructure using test containers:

```go
func TestPostgresUserRepo(t *testing.T) {
    container := testcontainers.NewPostgres(t)
    repo := postgres.NewUserRepository(container.DB)

    // Test against real database
    err := repo.Create(ctx, user)
    assert.NoError(t, err)
}
```

### What to Test Where

| Layer | Test Type | What to Mock |
|---|---|---|
| Domain | Unit tests | Nothing — pure logic |
| App (commands/queries) | Unit tests | Port interfaces |
| Adapters | Integration tests | Nothing — use real infra |
| Handlers | Unit tests | Application struct |
