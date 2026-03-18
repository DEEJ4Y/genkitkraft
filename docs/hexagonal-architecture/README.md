# Hexagonal Architecture Guide

> **Hexagonal Architecture** (also known as Ports & Adapters) organizes code so that business logic is at the center, isolated from all external concerns. The core application communicates with the outside world only through well-defined interfaces (ports), and concrete implementations (adapters) plug into those interfaces.

## Table of Contents

1. [Core Principles](01-core-principles.md)
2. [Directory Structure](02-directory-structure.md)
3. [The Domain Layer](03-domain-layer.md)
4. [Ports — Defining Contracts](04-ports.md)
5. [Adapters — Implementing Contracts](05-adapters.md)
6. [The Application Layer (Use Cases)](06-application-layer.md)
7. [Handlers — Primary Adapters](07-handlers.md)
8. [Dependency Injection & Wiring](08-dependency-injection.md)
9. [Cross-Cutting Concerns with the Decorator Pattern](09-decorators.md)
10. [Error Handling Strategy](10-error-handling.md)
11. [Configuration Management](11-configuration.md)
12. [Testing Strategy](12-testing.md)
13. [Dependency Flow Rules](13-dependency-flow-rules.md)
