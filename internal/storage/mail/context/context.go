package context

import (
	"context"

	"github.com/emersion/go-imap/client"
	gctx "github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/storage"
)

type MailContextManager struct {
	key gctx.CtxKey
}

func New(key gctx.CtxKey) *MailContextManager {
	return &MailContextManager{key}
}

func (ctxman *MailContextManager) Set(ctx context.Context, val *client.Client) context.Context {
	return context.WithValue(ctx, ctxman.key, val)
}

func (ctxman *MailContextManager) Get(ctx context.Context) (*client.Client, error) {
	val := ctx.Value(ctxman.key)
	if val == nil {
		return nil, storage.ErrNotEnrichedContext
	}

	c, ok := val.(*client.Client)
	if !ok {
		return nil, storage.ErrInvalidValue
	}

	return c, nil
}
