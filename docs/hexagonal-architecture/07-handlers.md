# Handlers — Primary Adapters

Handlers are primary adapters that translate external protocols (HTTP, gRPC, CLI, events) into application layer calls.

## gRPC Handler Example

```go
// internal/handlers/grpc_service/admin_service.go

type AdminService struct {
    pb.UnimplementedAdminServiceServer
    app *app.AdminApp
}

func NewAdminService(app *app.AdminApp) *AdminService {
    return &AdminService{app: app}
}

func (s *AdminService) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
    // 1. Extract context (e.g., auth info set by interceptors)
    userID := ctx.Value(contextkey.UserID).(uint64)

    // 2. Map proto → application DTO
    id, err := s.app.Commands.CreateOrder.Execute(ctx, commands.CreateOrder{
        CustomerID: domain.ID(req.CustomerId),
        Items:      mapProtoItems(req.Items),
    })
    if err != nil {
        return nil, mapErrorToGRPCStatus(err)
    }

    // 3. Map result → proto response
    return &pb.CreateOrderResponse{Id: string(id)}, nil
}
```

## Event Worker Handler

```go
// internal/handlers/worker/processor.go

type Processor struct {
    app *app.WorkerApp
}

func (p *Processor) Process(ctx context.Context, event *event.Data) error {
    switch event.Type {
    case event.TypeOrderCancelled:
        _, err := p.app.Commands.CancelOrder.Execute(ctx, commands.CancelOrder{
            OrderID: domain.ID(event.OrderID),
        })
        return err
    }
    return nil
}
```

## Type Conversion (Handler-Level)

```go
// internal/handlers/grpc_service/type_conversion.go

func mapProtoItems(items []*pb.OrderItem) []commands.CreateOrderItem { ... }
func mapErrorToGRPCStatus(err error) error {
    var appErr *errors.AppError
    if errors.As(err, &appErr) {
        switch appErr.Code() {
        case errors.NotFound:
            return status.Error(codes.NotFound, appErr.Message())
        case errors.InvalidInput:
            return status.Error(codes.InvalidArgument, appErr.Message())
        // ...
        }
    }
    return status.Error(codes.Internal, "internal error")
}
```

## Interceptors / Middleware

Interceptors handle cross-cutting concerns at the transport layer:

```go
// internal/handlers/grpc_service/interceptors/

// Authentication — validates JWT, sets user context
func AuthInterceptor(jwtService JwtService) grpc.UnaryServerInterceptor

// Authorization — checks permissions via policy engine
func AuthzInterceptor(opaClient OPAClient, permission string) grpc.UnaryServerInterceptor

// Correlation ID — propagates request tracing
func CorrelationIDInterceptor() grpc.UnaryServerInterceptor

// Logging — logs request/response
func LoggingInterceptor(logger zerolog.Logger) grpc.UnaryServerInterceptor
```

## Rules

- Handlers import from `app/` and `domain/` — never from `adapters/`.
- Handlers are responsible for **protocol translation** only — no business logic.
- Each handler has its own `type_conversion.go` for mapping between transport types and application DTOs.
