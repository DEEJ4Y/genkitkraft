# Core Principles

1. **The domain is the center of the universe.** Business logic has zero knowledge of databases, HTTP frameworks, message queues, or any external technology.
2. **Dependencies point inward.** Outer layers depend on inner layers — never the reverse.
3. **Communication through ports.** The application core defines interfaces (ports) that describe what it needs. External systems provide implementations (adapters) for those interfaces.
4. **Swappable infrastructure.** Because all infrastructure is behind interfaces, you can replace MySQL with Postgres, Redis with Memcached, or REST with gRPC without touching business logic.

```
                ┌─────────────────────────────────────┐
                │          Primary Adapters            │
                │   (gRPC handlers, HTTP controllers,  │
                │    CLI, event consumers)             │
                └───────────────┬─────────────────────┘
                                │ calls
                ┌───────────────▼─────────────────────┐
                │        Application Layer             │
                │   (Commands, Queries, Use Cases)     │
                │                                      │
                │   Uses ports (interfaces) to talk     │
                │   to the outside world               │
                └───────────────┬─────────────────────┘
                                │ depends on
                ┌───────────────▼─────────────────────┐
                │          Domain Layer                │
                │   (Entities, Value Objects,           │
                │    Business Rules)                   │
                └───────────────┬─────────────────────┘
                                │ defined by
                ┌───────────────▼─────────────────────┐
                │            Ports                     │
                │   (Repository interfaces,            │
                │    Service interfaces)               │
                └───────────────┬─────────────────────┘
                                │ implemented by
                ┌───────────────▼─────────────────────┐
                │        Secondary Adapters            │
                │   (MySQL, Redis, S3, GraphQL         │
                │    clients, external APIs)           │
                └─────────────────────────────────────┘
```
