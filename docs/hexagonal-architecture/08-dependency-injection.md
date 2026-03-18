# Dependency Injection & Wiring

The composition root lives in `internal/services/` and is the **only place** that knows about all layers. It wires adapters into application use cases and application use cases into handlers.

```go
// internal/services/application.go

type ApplicationDependencies struct {
    ReadDB            *sql.DB
    WriteDB           *sql.DB
    RedisClient       *redis.Client
    NotificationClient notification.Client
    Logger            zerolog.Logger
}

func NewAdminApplication(deps ApplicationDependencies) *app.AdminApp {
    // 1. Instantiate adapters
    repo := mysql_repository.NewOrderMySQLRepository(deps.ReadDB, deps.WriteDB)
    notifSvc := notification_adapter.NewNotificationService(deps.NotificationClient)
    cacheStore := cache_adapter.NewCacheStore(deps.RedisClient)

    // 2. Create use cases with port implementations injected
    createOrderCmd := commands.NewCreateOrderCommand(repo, notifSvc, deps.Logger)
    listOrdersQuery := queries.NewListOrdersQuery(repo, deps.Logger)

    // 3. Apply decorators for cross-cutting concerns (see next section)
    // ...

    // 4. Return the application struct
    return &app.AdminApp{
        Commands: app.AdminCommands{
            CreateOrder: createOrderCmd,
        },
        Queries: app.AdminQueries{
            ListOrders: listOrdersQuery,
        },
    }
}
```

```go
// cmd/server/main.go

func main() {
    cfg := config.Load()

    // Initialize infrastructure
    readDB := database.Connect(cfg.ReadDB)
    writeDB := database.Connect(cfg.WriteDB)
    redisClient := cache.Connect(cfg.Cache)

    // Wire application
    adminApp := services.NewAdminApplication(services.ApplicationDependencies{
        ReadDB:     readDB,
        WriteDB:    writeDB,
        RedisClient: redisClient,
        Logger:     logger,
    })

    // Wire server
    server := services.NewGRPCServer(cfg.GRPC, adminApp)
    server.Start()
}
```

## Rules

- Use **manual dependency injection** — no magic frameworks needed.
- The composition root is the only package that imports from all layers.
- Main function (`cmd/`) constructs infrastructure, calls the composition root, and starts the server.
