package context

import (
	"context"
	"log/slog"

	"github.com/emersion/go-imap/client"
	gctx "github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
)

type MailContextManager struct {
	key gctx.CtxKey
	l   *slog.Logger
}

func New(key gctx.CtxKey) *MailContextManager {
	return &MailContextManager{
		key: key,
		l: slog.With(
			sl.Module("MailContextManager"),
			slog.Any("key", key),
		),
	}
}

func (ctxman *MailContextManager) Set(ctx context.Context, val *client.Client) context.Context {
	fn := "context.Set"
	log := ctxman.l.With(sl.Method(fn))

	log.Debug("setting context", slog.Any("val", val))
	return context.WithValue(ctx, ctxman.key, val)
}

func (ctxman *MailContextManager) Get(ctx context.Context) (*client.Client, error) {
	fn := "context.Get"
	log := ctxman.l.With(sl.Method(fn))

	val := ctx.Value(ctxman.key)
	if val == nil {
		log.Error("empty context")
		return nil, storage.ErrNotEnrichedContext
	}

	c, ok := val.(*client.Client)
	if !ok {
		return nil, storage.ErrInvalidValue
	}
	log.Debug("extracted connection from context", slog.Any("c", c))

	return c, nil
}
