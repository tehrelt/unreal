package smtp

import (
	"context"
	"fmt"

	gctx "github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/storage"
)

var _ storage.Manager = (*Manager)(nil)

type Manager struct {
	*smtpCtxManager
}

func NewManager() storage.Manager {
	return &Manager{defaultManager}
}

func (m *Manager) Do(ctx context.Context, fn func(ctx context.Context) error) error {

	u, ok := ctx.Value(gctx.CtxKeyUser).(*entity.SessionInfo)
	if !ok {
		return fmt.Errorf("no user in context")
	}

	nctx := m.set(ctx, u)

	return fn(nctx)
}
