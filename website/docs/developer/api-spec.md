---
sidebar_position: 2
---

# API Specification (TypeSpec)

GenKitKraft uses **TypeSpec** for API-first development. API contracts are defined in `.tsp` files that compile to OpenAPI, which then generates Go server stubs and TypeScript client code. This ensures the API contract is always the source of truth — implementation follows the spec, never the other way around.

> For the complete TypeSpec language reference, see [`docs/api-spec/01-typespec-guide.md`](https://github.com/DEEJ4Y/genkitkraft/tree/main/docs/api-spec/01-typespec-guide.md).

## Spec Location

All TypeSpec files live in `spec/`:

```
spec/
  main.tsp                          → Entry point, service metadata, imports routes
  models/
    <feature>.tsp                   → Data model definitions
  routes/
    <feature>.tsp                   → Route definitions using models
  tsp-output/schema/
    openapi.yaml                    → Generated OpenAPI output (do not edit)
```

## Development Workflow

**Every API change must follow this sequence.** Do not skip steps or write Go code before the spec is updated and stubs are generated.

```
 ┌──────────────────┐
 │ 1. Update .tsp   │  Define/update models and routes
 │    files         │
 └────────┬─────────┘
          ▼
 ┌──────────────────┐
 │ 2. make          │  TypeSpec → OpenAPI YAML
 │    generate-spec │
 └────────┬─────────┘
          ▼
 ┌──────────────────┐
 │ 3. make          │  OpenAPI → Go server interface + types
 │    generate-go   │
 └────────┬─────────┘
          ▼
 ┌──────────────────┐
 │ 4. make          │  OpenAPI → TypeScript API client
 │    generate-ts   │
 └────────┬─────────┘
          ▼
 ┌──────────────────┐
 │ 5. Implement     │  Write handlers, app logic, ports, adapters
 │    handlers      │
 └──────────────────┘
```

**Shortcut:** Run `make generate` to execute steps 2–4 in one command.

## Project Conventions

### Namespace Structure

| Namespace | Purpose |
|---|---|
| `Api` | Top-level service namespace |
| `Api.Models` | Shared data model definitions |
| `Api.Routes` | Route/operation definitions |

### Conventions

- Group routes under `@tag("<Feature>")` namespaces for OpenAPI grouping
- Use `@summary()` and JSDoc comments for documentation
- Define response models explicitly in `models/`
- Routes import models via `using Api.Models;`
- New route files must be imported in `spec/main.tsp`

## Quick Patterns

### Defining a Model with Enum

```typespec
// spec/models/thing.tsp
namespace Api.Models;

enum ThingStatus {
  Active: "active",
  Inactive: "inactive",
}

model Thing {
  id: string;
  name: string;
  status: ThingStatus;
  createdAt: utcDateTime;
}

model CreateThingRequest {
  name: string;
  status?: ThingStatus;
}
```

### Defining CRUD Routes

```typespec
// spec/routes/things.tsp
import "@typespec/http";
import "../models/thing.tsp";

using TypeSpec.Http;
using Api.Models;

@tag("Things")
namespace Api.Routes.Things {

  @get @route("/things") @summary("List things")
  op list(): Thing[];

  @get @route("/things/{id}") @summary("Get thing")
  op get(@path id: string): Thing | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  @post @route("/things") @summary("Create thing")
  op create(@body body: CreateThingRequest): {
    @statusCode statusCode: 201;
    @body body: Thing;
  };

  @delete @route("/things/{id}") @summary("Delete thing")
  op delete(@path id: string): {
    @statusCode statusCode: 204;
  };
}
```

### Common Type Patterns

```typespec
name?: string;              // Optional field
value: string | null;       // Nullable field
value?: string | null;      // Optional + nullable
items: Item[];              // Array
metadata: Record<string>;   // Map / dictionary
status: "a" | "b" | "c";   // Literal union
model B { ...A; extra: string; }   // Spread (mixin)
model B extends A { extra: string; } // Inheritance
```

## Key Rules

:::warning Critical Rules
1. **Never manually edit generated files** — `internal/api/gen/*.gen.go` and `spec/tsp-output/` are overwritten on each generation.
2. **Always re-run generation before updating Go code** — Run `make generate` whenever the spec changes, before writing any implementation.
3. **`ServerInterface` is the source of truth** — The generated interface in `internal/api/gen/server.gen.go` defines handler signatures. Never hand-write route registrations.
4. **Handlers implement `gen.ServerInterface`** — The composition root registers them via `gen.HandlerFromMux()`.
:::
