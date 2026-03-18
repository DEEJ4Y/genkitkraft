# Testing Strategy

Hexagonal architecture makes testing straightforward because every dependency is behind an interface.

## Unit Tests (Application Layer)

Test use cases by providing mock implementations of ports:

```
internal/app/commands/create_order_test.go     # Mock the repository port
internal/app/queries/list_orders_test.go        # Mock the repository port
```

```go
func TestCreateOrder_Success(t *testing.T) {
    mockRepo := &MockOrderRepository{
        CreateFn: func(ctx context.Context, params repository.CreateOrderParams) (domain.ID, error) {
            return "order-123", nil
        },
    }

    cmd := &CreateOrderCommand{repo: mockRepo}
    id, err := cmd.Execute(ctx, CreateOrder{CustomerID: "cust-1"})

    assert.NoError(t, err)
    assert.Equal(t, domain.ID("order-123"), id)
}
```

## Integration Tests (Adapter Layer)

Test adapters against real infrastructure using test containers:

```
internal/adapters/mysql_repository/repository_test.go    # Real MySQL via testcontainers
internal/adapters/cache_store/cache_store_test.go         # Real Redis via testcontainers
```

## Test Infrastructure

```
resources/test/
├── containers/          # Testcontainer definitions (MySQL, Redis, Kafka, etc.)
│   ├── mysql.go
│   └── redis.go
├── seed/                # Database fixtures
│   └── seed.go
└── mock/                # Mock implementations of external services
    ├── jwt_service.go
    └── notification.go
```

## Rules

- **Unit tests** mock ports — fast, no infrastructure needed.
- **Integration tests** use test containers for real infrastructure — slower, validates adapter correctness.
- Mocks live in a shared `resources/test/mock/` directory.
- Compile-time interface checks (`var _ Port = (*Mock)(nil)`) ensure mocks stay in sync with ports.
