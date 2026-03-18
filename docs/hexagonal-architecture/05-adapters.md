# Adapters — Implementing Contracts

Adapters provide **concrete implementations** of port interfaces. Each adapter lives in its own package under `internal/adapters/`.

## Secondary Adapters (Outbound)

```go
// internal/adapters/mysql_repository/repository.go

type OrderMySQLRepository struct {
    readDB  *sql.DB
    writeDB *sql.DB
}

func NewOrderMySQLRepository(readDB, writeDB *sql.DB) *OrderMySQLRepository {
    return &OrderMySQLRepository{readDB: readDB, writeDB: writeDB}
}

// Compile-time interface check
var _ repository.OrderRepository = (*OrderMySQLRepository)(nil)

func (r *OrderMySQLRepository) Create(ctx context.Context, params repository.CreateOrderParams) (domain.ID, error) {
    tx, err := r.writeDB.BeginTx(ctx, nil)
    if err != nil {
        return "", err
    }
    defer tx.Rollback()

    // Execute queries, map params to SQL, handle transactions
    id := generateID()
    _, err = tx.ExecContext(ctx, "INSERT INTO orders ...", id, params.CustomerID)
    if err != nil {
        return "", err
    }

    return domain.ID(id), tx.Commit()
}
```

## Type Conversion

Adapters contain their own mapping logic to translate between port DTOs and infrastructure-specific types:

```go
// internal/adapters/mysql_repository/type_conversion.go

func toGetByIDResult(row sqlc.OrderRow) repository.GetByIDResult {
    return repository.GetByIDResult{
        ID:        domain.ID(row.ID),
        Status:    domain.OrderStatus(row.Status),
        CreatedAt: row.CreatedAt,
    }
}
```

## Rules

- Adapters import from `ports/` and `domain/` — never from `app/` or other adapters.
- Each adapter has its own `type_conversion.go` for mapping between layers.
- Use **compile-time interface checks** (`var _ Port = (*Adapter)(nil)`) to ensure adapters satisfy their port.
- Infrastructure concerns (connection pooling, retries, circuit breakers) live in the adapter, not in the port.
