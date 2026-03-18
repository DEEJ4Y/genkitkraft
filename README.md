# Your Project Name Here

Monorepo template for vibe coding projects. Frontend and backend services live together with full documentation, so coding agents have the context they need.

## Why

- All code and docs in one repo = better agent context
- Hexagonal architecture keeps things structured as complexity grows
- API contracts defined in TypeSpec, generating OpenAPI specs, server stubs, and client SDKs

## Docs

Everything lives under `docs/`:

- [Hexagonal Architecture Guide](docs/hexagonal-architecture/README.md) — project structure, patterns, dependency rules
- [TypeSpec Guide](docs/api-spec/01-typespec-guide.md) — API contract definitions

API spec implementations and generated OpenAPI output are in `spec/`.

## Base Spec

Health check server with two endpoints:

- `GET /readyz` — readiness probe
- `GET /livez` — liveness probe
