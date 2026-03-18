# Ports — Defining Contracts

Ports are **interfaces** that define the contracts between the application core and the outside world. They live in their own package under `internal/ports/`.

## Outbound Ports (Driven)

These describe what the application **needs from external systems**:

```go
// internal/ports/repository/interface.go

type OrderRepository interface {
    Create(ctx context.Context, params CreateOrderParams) (domain.ID, error)
    GetByID(ctx context.Context, params GetByIDParams) (GetByIDResult, error)
    List(ctx context.Context, params ListParams) (ListResult, error)
    Delete(ctx context.Context, params DeleteParams) error
}
```

```go
// internal/ports/notification_service/interface.go

type NotificationService interface {
    SendOrderConfirmation(ctx context.Context, params SendConfirmationParams) error
}
```

## Port-Specific DTOs

Each port defines its own parameter and result types. This prevents domain types from leaking infrastructure details:

```go
// internal/ports/repository/types.go

type CreateOrderParams struct {
    CustomerID domain.ID
    Items      []CreateOrderItemParams
}

type GetByIDResult struct {
    ID        domain.ID
    Status    domain.OrderStatus
    Items     []ItemResult
    CreatedAt time.Time
}
```

## Inbound Ports (Driving)

These describe how external triggers enter the application:

```go
// internal/ports/event/event.go

type Consumer interface {
    Start(ctx context.Context, handler Handler) error
    Stop() error
}

type Handler func(ctx context.Context, data *EventData) error
```

## Rules

- Port interfaces should be **technology-agnostic** — no SQL types, no HTTP types, no protobuf types.
- Each port package contains its interface and its DTOs.
- Port DTOs can reference domain types (e.g., `domain.ID`) but never infrastructure types.
