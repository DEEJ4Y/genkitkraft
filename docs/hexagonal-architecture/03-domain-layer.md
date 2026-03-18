# The Domain Layer

The domain layer contains your business entities, value objects, and business rules. It has **no dependencies on any other package** — it is the innermost layer.

## Entities

```go
// internal/domain/order.go

type Order struct {
    ID        ID
    Items     []Item
    Status    OrderStatus
    CreatedAt time.Time
    Details
}

type Item struct {
    ID       ID
    Name     string
    Quantity int
    Price    Money
}
```

## Value Objects

Use custom types to add semantic meaning and validation:

```go
// internal/domain/types.go

type ID string
type OrderStatus string
type Money int64  // cents

const (
    StatusPending   OrderStatus = "pending"
    StatusConfirmed OrderStatus = "confirmed"
    StatusShipped   OrderStatus = "shipped"
)

func (s OrderStatus) IsValid() bool {
    switch s {
    case StatusPending, StatusConfirmed, StatusShipped:
        return true
    }
    return false
}
```

## Rules

- Domain types are **pure data** — no database tags, no JSON annotations (unless truly domain-relevant).
- Domain logic lives in methods on entities or standalone functions within the domain package.
- The domain package **never imports** from `ports/`, `adapters/`, `app/`, or `handlers/`.
