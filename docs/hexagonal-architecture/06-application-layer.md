# The Application Layer (Use Cases)

The application layer orchestrates domain logic by composing port calls into meaningful use cases. It is split into **commands** (writes) and **queries** (reads).

## The Executor Pattern

Define a generic executor interface so all use cases share a common shape:

```go
// internal/app/executors/executor.go

type Executor[Params any] interface {
    Execute(ctx context.Context, params Params) error
}

type ExecutorWithReturn[Params, Result any] interface {
    Execute(ctx context.Context, params Params) (Result, error)
}
```

## Commands (Write Operations)

```go
// internal/app/commands/create_order.go

type CreateOrderCommand struct {
    repo              repository.OrderRepository
    notificationSvc   notification_service.NotificationService
}

// Input DTO for this command
type CreateOrder struct {
    CustomerID domain.ID
    Items      []CreateOrderItem
}

func (c *CreateOrderCommand) Execute(ctx context.Context, params CreateOrder) (domain.ID, error) {
    // 1. Validate input
    if err := params.validate(); err != nil {
        return "", err
    }

    // 2. Execute business logic (delegates to repository port)
    id, err := c.repo.Create(ctx, repository.CreateOrderParams{
        CustomerID: params.CustomerID,
        Items:      mapItems(params.Items),
    })
    if err != nil {
        return "", err
    }

    // 3. Side effects (delegates to notification port)
    _ = c.notificationSvc.SendOrderConfirmation(ctx, ...)

    return id, nil
}
```

## Queries (Read Operations)

```go
// internal/app/queries/get_order.go

type GetOrderQuery struct {
    repo repository.OrderRepository
}

type GetOrder struct {
    OrderID domain.ID
}

type GetOrderResult struct {
    ID     domain.ID
    Status domain.OrderStatus
    Items  []OrderItem
}

func (q *GetOrderQuery) Execute(ctx context.Context, params GetOrder) (GetOrderResult, error) {
    result, err := q.repo.GetByID(ctx, repository.GetByIDParams{ID: params.OrderID})
    if err != nil {
        return GetOrderResult{}, err
    }
    return mapToGetOrderResult(result), nil
}
```

## Application Struct

The application struct groups all commands and queries into a single object that handlers receive:

```go
// internal/app/admin_app.go

type AdminApp struct {
    Commands AdminCommands
    Queries  AdminQueries
}

type AdminCommands struct {
    CreateOrder ExecutorWithReturn[CreateOrder, domain.ID]
    DeleteOrder Executor[DeleteOrder]
}

type AdminQueries struct {
    GetOrder   ExecutorWithReturn[GetOrder, GetOrderResult]
    ListOrders ExecutorWithReturn[ListOrders, ListOrdersResult]
}
```

## Rules

- Commands and queries depend only on **port interfaces**, never on concrete adapters.
- Each command/query has its own input DTO with a `validate()` method.
- The application layer **does not know** about HTTP, gRPC, or any transport mechanism.
