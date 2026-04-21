package httptoolrepo

import (
	"context"

	httptool "github.com/DEEJ4Y/genkitkraft/internal/domain/http_tool"
)

// HttpToolRepository defines the contract for HTTP tool persistence.
type HttpToolRepository interface {
	List(ctx context.Context, limit, offset int) ([]*httptool.HttpTool, error)
	Count(ctx context.Context) (int, error)
	GetByID(ctx context.Context, id string) (*httptool.HttpTool, error)
	Create(ctx context.Context, t *httptool.HttpTool) error
	Update(ctx context.Context, t *httptool.HttpTool) error
	Delete(ctx context.Context, id string) error
}
