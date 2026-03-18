# Dependency Flow Rules

This is the most important section. If you follow these rules, your architecture will remain clean:

```
✅ ALLOWED                          ❌ FORBIDDEN
─────────────────────────           ─────────────────────────
domain     → (nothing)              domain     → ports
ports      → domain                 ports      → adapters
app        → ports, domain          app        → adapters, handlers
adapters   → ports, domain          adapters   → app, handlers, other adapters
handlers   → app, domain            handlers   → adapters
services   → everything             (services is the composition root)
```

## Package Import Rules

| Package | Can import from | Cannot import from |
|---------|----------------|-------------------|
| `domain/` | standard library only | everything else |
| `ports/` | `domain/` | `adapters/`, `app/`, `handlers/` |
| `app/` | `ports/`, `domain/` | `adapters/`, `handlers/` |
| `adapters/` | `ports/`, `domain/`, `clients/` | `app/`, `handlers/` |
| `handlers/` | `app/`, `domain/`, `common/` | `adapters/` |
| `services/` | all internal packages | — |
| `cmd/` | `services/`, `config/` | — |

## Quick Validation

If you ever find yourself importing an adapter package inside `app/`, or importing `app/` inside an adapter — **stop**. You're violating the architecture. Instead:

1. Define an interface (port) for what you need.
2. Implement it in an adapter.
3. Inject it through the composition root.

## Summary

| Concept | Location | Responsibility |
|---------|----------|---------------|
| **Domain** | `internal/domain/` | Pure business entities and rules |
| **Ports** | `internal/ports/` | Interfaces defining external contracts |
| **Adapters** | `internal/adapters/` | Concrete implementations (DB, cache, APIs) |
| **Application** | `internal/app/` | Use cases orchestrating domain + ports |
| **Handlers** | `internal/handlers/` | Protocol translation (gRPC, HTTP, events) |
| **Decorators** | `internal/app/decorators/` | Cross-cutting concerns (logging, tracing, caching) |
| **Services** | `internal/services/` | Composition root — wires everything together |
| **Config** | `internal/config/` | Configuration structs and loading |
| **Clients** | `internal/clients/` | Low-level infrastructure client wrappers |

The power of hexagonal architecture lies in its constraints: by enforcing strict dependency rules and communicating through interfaces, you gain the ability to test, swap, and evolve each layer independently.
