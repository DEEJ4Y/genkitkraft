# Directory Structure

```
project-root/
├── cmd/                          # Entry points
│   ├── server/                   # Main server (HTTP/gRPC)
│   │   └── main.go
│   └── worker/                   # Background worker
│       └── main.go
│
├── internal/                     # Private application code
│   ├── domain/                   # Core business entities & rules
│   │   ├── entity.go
│   │   ├── value_objects.go
│   │   └── rules.go
│   │
│   ├── ports/                    # Interface definitions (contracts)
│   │   ├── repository/           # Persistence port
│   │   │   ├── interface.go      # Interface definition
│   │   │   └── types.go          # Port-specific DTOs (params, results)
│   │   ├── cache_store/          # Caching port
│   │   │   └── interface.go
│   │   ├── external_service/     # External service port
│   │   │   ├── interface.go
│   │   │   └── types.go
│   │   ├── file_store/           # File storage port
│   │   │   └── interface.go
│   │   └── event/                # Event consumption port
│   │       └── interface.go
│   │
│   ├── adapters/                 # Concrete implementations of ports
│   │   ├── mysql_repository/     # Database adapter
│   │   │   ├── repository.go
│   │   │   ├── type_conversion.go
│   │   │   └── repository_test.go
│   │   ├── cache_store/          # Redis adapter
│   │   │   └── cache_store.go
│   │   ├── external_service/     # HTTP/GraphQL client adapter
│   │   │   └── service.go
│   │   └── file_store/           # S3/Blob adapter
│   │       └── store.go
│   │
│   ├── app/                      # Application layer (use cases)
│   │   ├── commands/             # Write operations
│   │   │   ├── create_entity.go
│   │   │   └── delete_entity.go
│   │   ├── queries/              # Read operations
│   │   │   ├── list_entities.go
│   │   │   └── get_entity.go
│   │   ├── decorators/           # Cross-cutting concern wrappers
│   │   │   ├── logging.go
│   │   │   ├── tracing.go
│   │   │   ├── cache_invalidation.go
│   │   │   └── error_handler.go
│   │   ├── executors/            # Executor interfaces
│   │   │   └── executor.go
│   │   └── admin_app.go          # Application struct (wires commands & queries)
│   │
│   ├── handlers/                 # Primary adapters (entry points)
│   │   ├── grpc_service/         # gRPC handlers
│   │   │   ├── admin_service.go
│   │   │   ├── type_conversion.go
│   │   │   └── interceptors/     # Middleware (auth, logging, etc.)
│   │   └── worker/               # Event-driven handlers
│   │       └── processor.go
│   │
│   ├── clients/                  # Low-level infrastructure clients
│   │   ├── database/
│   │   ├── cache/
│   │   └── blob/
│   │
│   ├── common/                   # Shared utilities
│   │   ├── errors/               # Custom error types
│   │   ├── logger/               # Logging utilities
│   │   └── metrics/              # Observability
│   │
│   ├── config/                   # Configuration structs & loading
│   │   └── config.go
│   │
│   └── services/                 # Composition root (dependency injection)
│       ├── application.go        # Wires adapters → app layer
│       └── server.go             # Wires handlers → server
│
├── resources/
│   └── test/                     # Test infrastructure
│       ├── containers/           # Test containers (DB, Redis, etc.)
│       ├── seed/                 # Database fixtures
│       └── mock/                 # Mocks for external services
│
└── proto/                        # Protocol buffer definitions (if gRPC)
```

## Why This Structure Works

| Directory | Layer | Responsibility |
|-----------|-------|---------------|
| `domain/` | Core | Pure business logic, no imports from other internal packages |
| `ports/` | Core boundary | Interfaces that define what the application needs |
| `app/` | Application | Orchestrates domain logic through use cases |
| `adapters/` | Infrastructure | Implements ports with real technology |
| `handlers/` | Infrastructure | Translates external requests into application calls |
| `services/` | Composition | Wires everything together (the only place that knows all layers) |
