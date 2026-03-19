---
name: dev-flow
description: Guidelines and rules for developing in the GenKitKraft codebase, including architecture principles, package structure, API spec conventions, and testing practices.
---

# GenKitKraft Development Skill

## When to Use

Invoke this skill whenever building, modifying, or reviewing code in this project. This includes adding features, creating new endpoints, writing domain logic, defining ports/adapters, or modifying API specs.

## Project Overview

Self-hostable platform for configuring and running LLM agents. Built on Google Genkit (Go SDK). Provides an agent builder UI, MCP tool support, multi-provider LLM access, and an OpenAI-compatible API.

Tech stack: Go (backend, hexagonal architecture), SQLite (default storage), TypeSpec (API contract definitions), embedded frontend in single binary.

## Architecture Rules (Hexagonal / Ports & Adapters)

Before writing any Go code, consult `docs/hexagonal-architecture/` for the full guide. The critical rules are:

### Dependency Flow (MUST follow)

```
ALLOWED                              FORBIDDEN
───────────────────────              ───────────────────────
domain     → (nothing)               domain     → ports
ports      → domain                  ports      → adapters
app        → ports, domain           app        → adapters, handlers
adapters   → ports, domain           adapters   → app, handlers, other adapters
handlers   → app, domain             handlers   → adapters
services   → everything              (services is the composition root)
```

If you find yourself importing an adapter inside `app/`, or `app/` inside an adapter — stop. Define a port interface, implement it in an adapter, inject through the composition root.

### Package Import Reference

| Package     | Can import                      | Cannot import                    |
| ----------- | ------------------------------- | -------------------------------- |
| `domain/`   | standard library only           | everything else                  |
| `ports/`    | `domain/`                       | `adapters/`, `app/`, `handlers/` |
| `app/`      | `ports/`, `domain/`             | `adapters/`, `handlers/`         |
| `adapters/` | `ports/`, `domain/`, `clients/` | `app/`, `handlers/`              |
| `handlers/` | `app/`, `domain/`, `common/`    | `adapters/`                      |
| `services/` | all internal packages           | —                                |
| `cmd/`      | `services/`, `config/`          | —                                |

### Directory Structure

```
cmd/                    → Entry points (main.go)
internal/
  domain/               → Pure business entities, value objects, rules (NO external imports)
  ports/                → Interface definitions with port-specific DTOs
    <port_name>/
      interface.go      → Interface definition
      types.go          → Port-specific param/result DTOs
  adapters/             → Concrete implementations of ports
    <adapter_name>/
      <impl>.go         → Implementation
      type_conversion.go → Mapping between port DTOs and infra types
  app/                  → Application layer (use cases)
    commands/           → Write operations (each file = one command)
    queries/            → Read operations (each file = one query)
    decorators/         → Cross-cutting wrappers (logging, tracing, caching, errors)
    executors/          → Generic Executor interfaces
    <name>_app.go       → Application struct grouping commands & queries
  handlers/             → Primary adapters (HTTP/gRPC/CLI/event translation)
    <handler_name>/
      <service>.go
      type_conversion.go
      interceptors/     → Middleware (auth, logging, correlation ID)
  clients/              → Low-level infrastructure client wrappers
  common/               → Shared utilities (errors, logger, metrics)
  config/               → Configuration structs, loaded from env vars
  services/             → Composition root (dependency injection wiring)
resources/test/         → Test infrastructure (containers, seed, mocks)
```

### Build & Generation Commands

| Command | What it does |
| --- | --- |
| `make generate` | Full pipeline: TypeSpec → OpenAPI → Go stubs → TS client |
| `make generate-spec` | Compile TypeSpec to OpenAPI YAML |
| `make generate-go` | Generate Go server interface + types from OpenAPI |
| `make generate-ts` | Generate TypeScript API client from OpenAPI |
| `make build` | Build the Go server binary |

### Key Patterns

**Executor pattern** — All use cases implement a generic interface:

```go
type Executor[Params any] interface {
    Execute(ctx context.Context, params Params) error
}
type ExecutorWithReturn[Params, Result any] interface {
    Execute(ctx context.Context, params Params) (Result, error)
}
```

**Application struct** — Groups commands and queries into a single injectable object:

```go
type AdminApp struct {
    Commands AdminCommands
    Queries  AdminQueries
}
```

**Decorator pattern** — Wrap executors for logging, tracing, caching, error handling. Applied in the composition root. Order matters (outermost executes first).

**Compile-time interface checks** — Every adapter must include:

```go
var _ portpkg.SomeInterface = (*AdapterImpl)(nil)
```

**Type conversion** — Each adapter and handler has its own `type_conversion.go`. Never leak infrastructure types into ports or domain.

**Manual dependency injection** — No DI frameworks. The composition root (`internal/services/`) is the only place that knows all layers.

**Error handling** — Use `AppError` with error codes (`NotFound`, `InvalidInput`, `Conflict`, etc.) from `internal/common/errors/`. Handlers map codes to transport status codes. The error handler decorator wraps unexpected errors as `Internal`.

**Configuration** — All config from environment variables, loaded once at startup, injected through composition root. Adapters receive only what they need.

### Testing

- **Unit tests**: Mock port interfaces for app layer tests. Fast, no infra needed.
- **Integration tests**: Test adapters against real infra via test containers.
- **Mocks**: Live in `resources/test/mock/`. Use compile-time interface checks.

## API Specification (TypeSpec)

Before adding or modifying API endpoints, consult `docs/api-spec/01-typespec-guide.md` for the full TypeSpec reference.

### Spec Location

All TypeSpec files live in `spec/`:

- `spec/main.tsp` — Entry point, service metadata, imports routes
- `spec/models/` — Data model definitions
- `spec/routes/` — Route definitions using models
- `spec/tsp-output/schema/openapi.yaml` — Generated OpenAPI output

### Spec-Driven Development Workflow (MUST follow)

Every API change follows this strict sequence. Do NOT skip steps or implement code before the spec is updated and stubs are generated.

1. **Update API spec** — Define/update models in `spec/models/<feature>.tsp` and routes in `spec/routes/<feature>.tsp`. Import new route files in `spec/main.tsp`.
2. **Generate OpenAPI from TypeSpec** — Run `make generate-spec` (compiles TypeSpec to `spec/tsp-output/schema/openapi.yaml`).
3. **Generate server/client stubs** — Run `make generate-go` (generates `internal/api/gen/server.gen.go` and `types.gen.go` from OpenAPI) and `make generate-ts` (generates TypeScript client in `ui/`). Or run `make generate` to do all three steps at once.
4. **Update implementations** — Implement the corresponding Go handler, app commands/queries, ports, and adapters to satisfy the newly generated `ServerInterface`.

**Key rules:**
- The generated `ServerInterface` in `internal/api/gen/server.gen.go` is the source of truth for HTTP handler signatures. Never hand-write route registrations.
- Generated files (`internal/api/gen/*.gen.go`) must NEVER be manually edited. They are overwritten on each generation.
- When modifying existing endpoints, always re-run the full generation pipeline (`make generate`) before updating Go code, so generated types stay in sync.
- Handlers in `internal/handlers/` implement `gen.ServerInterface`. The composition root registers them via `gen.HandlerFromMux()`.

### TypeSpec Conventions (this project)

- Service namespace: `Api`
- Models namespace: `Api.Models`
- Routes namespace: `Api.Routes`
- Group routes under `@tag("<Feature>")` namespaces
- Use `@summary()` and JSDoc comments for documentation
- Define response models explicitly in `models/`
- Routes import models and use `using Api.Models;`

### TypeSpec Quick Patterns

```typespec
// Model with enum
enum Status { Active: "active", Inactive: "inactive" }
model Thing { id: string; name: string; status: Status; }

// CRUD routes
@tag("Things")
namespace Things {
  @get @route("/things") @summary("List things")
  op list(): Thing[];

  @get @route("/things/{id}") @summary("Get thing")
  op get(@path id: string): Thing | { @statusCode statusCode: 404; @body body: ErrorResponse; };

  @post @route("/things") @summary("Create thing")
  op create(@body body: CreateThingRequest): { @statusCode statusCode: 201; @body body: Thing; };

  @delete @route("/things/{id}") @summary("Delete thing")
  op delete(@path id: string): { @statusCode statusCode: 204; };
}
```

## Checklist for New Features

### Phase 1: Spec-Driven Contract (do this FIRST, before any Go code)

1. [ ] Define API contract in TypeSpec (`spec/models/<feature>.tsp` + `spec/routes/<feature>.tsp`)
2. [ ] Import new route file in `spec/main.tsp` if it's a new file
3. [ ] Run `make generate` to compile spec → generate OpenAPI → generate Go server stubs + TS client
4. [ ] Verify the generated `ServerInterface` in `internal/api/gen/server.gen.go` has the new methods

### Phase 2: Hexagonal Implementation (follow dependency flow strictly)

5. [ ] Add domain entities/value objects in `internal/domain/` (stdlib only, no external deps)
6. [ ] Define port interfaces and DTOs in `internal/ports/<name>/` (imports domain only)
7. [ ] Implement adapters in `internal/adapters/<name>/` with compile-time checks (imports ports + domain)
8. [ ] Create commands/queries in `internal/app/commands/` or `internal/app/queries/` (imports ports + domain)
9. [ ] Add decorators if needed in `internal/app/decorators/` (imports app + executors)
10. [ ] Add handler with `type_conversion.go` in `internal/handlers/<name>/` (imports app + gen + common)
11. [ ] Wire everything in `internal/services/` composition root (imports all layers)

### Phase 3: Verification

12. [ ] Run `go build ./...` and `go vet ./...`
13. [ ] Write unit tests (mock port interfaces) and integration tests (test containers)
14. [ ] Verify dependency flow rules: no forbidden imports between layers

## Additional Resources

- [Hexagonal Architecture Guide](docs/hexagonal-architecture/README.md) - project structure, patterns, dependency rules
- [TypeSpec Guide](docs/api-spec/01-typespec-guide.md) - API contract definitions
