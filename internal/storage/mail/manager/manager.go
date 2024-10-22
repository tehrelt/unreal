package manager

import (
	"context"
	"fmt"

	gctx "github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/imap"
	mctx "github.com/tehrelt/unreal/internal/storage/mail/context"
)

type MailManager struct {
	ctxManager *mctx.MailContextManager
}

func New(key gctx.CtxKey) *MailManager {
	return &MailManager{
		ctxManager: mctx.New(key),
	}
}

func (m *MailManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {

	creds, ok := ctx.Value(gctx.CtxKeyUser).(*entity.SessionInfo)
	if !ok {
		return fmt.Errorf("no user in context")
	}

	c, cleanup, err := imap.Dial(creds.Email, creds.Password, creds.Imap.Host, creds.Imap.Port)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer cleanup()

	nctx := m.ctxManager.Set(ctx, c)

	return fn(nctx)
}
