---
sidebar_position: 3
---

# Contributing

## Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| Go | 1.25+ | Backend server |
| Node.js | 22+ | TypeSpec compiler, frontend |
| TypeSpec CLI | via `npm install` | API spec compilation |
| Docker | Latest | Integration tests (test containers) |

Install TypeSpec dependencies:

```bash
cd spec && npm install
```

## Project Structure

```
cmd/                → Entry points (main.go)
internal/           → All Go source code
  domain/           →   Pure business entities & value objects
  ports/            →   Interface definitions (contracts)
  adapters/         →   Port implementations (DB, external APIs)
  app/              →   Use cases (commands & queries)
  handlers/         →   HTTP handlers (implement gen.ServerInterface)
  services/         →   Composition root (dependency injection)
  common/           →   Shared utilities (errors, logger)
  config/           →   Configuration (env vars)
  clients/          →   Infrastructure client wrappers
spec/               → TypeSpec API definitions
ui/                 → Next.js frontend (React + Mantine)
website/            → Docusaurus documentation site
docs/               → In-repo developer documentation (detailed)
```

## Build Commands

| Command | Description |
|---|---|
| `make generate` | Full pipeline: TypeSpec → OpenAPI → Go stubs → TS client |
| `make generate-spec` | Compile TypeSpec to OpenAPI YAML |
| `make generate-go` | Generate Go server interface + types from OpenAPI |
| `make generate-ts` | Generate TypeScript API client from OpenAPI |
| `make build` | Build the Go server binary |

## Adding a New Feature

Follow this checklist in order. See [Architecture](./architecture.md) for layer details and [API Specification](./api-spec.md) for TypeSpec patterns.

### Phase 1: Define the API Contract

1. **Define models** in `spec/models/<feature>.tsp`
2. **Define routes** in `spec/routes/<feature>.tsp`
3. **Import** the new route file in `spec/main.tsp` (if new)
4. **Run `make generate`** to produce OpenAPI → Go stubs → TS client
5. **Verify** the generated `ServerInterface` in `internal/api/gen/server.gen.go` has the new methods

### Phase 2: Implement (Follow Dependency Flow)

6. **Add domain entities** in `internal/domain/` (stdlib only, no external deps)
7. **Define port interfaces** and DTOs in `internal/ports/<name>/`
8. **Implement adapters** in `internal/adapters/<name>/` with compile-time interface checks
9. **Create commands/queries** in `internal/app/commands/` or `internal/app/queries/`
10. **Add handler** with `type_conversion.go` in `internal/handlers/<name>/`
11. **Wire everything** in `internal/services/` (composition root)

### Phase 3: Verify

12. **Build:** `go build ./...` and `go vet ./...`
13. **Test:** Write unit tests (mock ports) and integration tests (test containers)
14. **Check imports:** Verify no forbidden cross-layer imports (see [dependency flow rules](./architecture.md#dependency-flow-rules))

## Frontend Development

The frontend is a Next.js app using React and Mantine UI:

```bash
cd ui
npm ci              # Install dependencies
npm run dev         # Start dev server
npm run generate:api  # Regenerate API client from OpenAPI
npm run build       # Production build
```

The API client is auto-generated from the OpenAPI spec. After any API changes, run `make generate-ts` (or `npm run generate:api` from the `ui/` directory) to keep the client in sync.

## Documentation

The documentation website uses [Docusaurus](https://docusaurus.io/) and lives in `website/`:

```bash
cd website
npm ci
npm run start       # Local dev server
npm run build       # Production build
```

- **Website docs** (user-facing): `website/docs/`
- **In-repo docs** (detailed developer reference): `docs/`

When adding features, update both the relevant in-repo docs and any affected website pages.
