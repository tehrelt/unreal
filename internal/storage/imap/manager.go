package imap

import (
	"context"
	"fmt"

	gctx "github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/imap"
	"github.com/tehrelt/unreal/internal/storage"
)

var _ storage.Manager = (*Manager)(nil)

type Manager struct {
	*imapCtxManager
}

func NewManager() storage.Manager {
	return &Manager{defaultManager}
}

func (m *Manager) Do(ctx context.Context, fn func(ctx context.Context) error) error {

	creds, ok := ctx.Value(gctx.CtxKeyUser).(*entity.SessionInfo)
	if !ok {
		return fmt.Errorf("no user in context")
	}

	c, cleanup, err := imap.Dial(creds.Email, creds.Password, creds.Imap.Host, creds.Imap.Port)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer cleanup()

	nctx := m.set(ctx, c)

	return fn(nctx)
}
