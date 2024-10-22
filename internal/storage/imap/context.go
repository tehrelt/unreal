package imap

import (
	"context"
	"log/slog"

	"github.com/emersion/go-imap/client"
	gctx "github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
)

var defaultManager = newCtxManager()

type imapCtxManager struct {
	key gctx.CtxKey
	l   *slog.Logger
}

func newCtxManager() *imapCtxManager {
	return &imapCtxManager{
		key: connKey,
		l: slog.With(
			sl.Module("imapCtxManager"),
			slog.Any("key", connKey),
		),
	}
}

func (ctxman *imapCtxManager) set(ctx context.Context, val *client.Client) context.Context {
	fn := "context.Set"
	log := ctxman.l.With(sl.Method(fn))

	log.Debug("setting context", slog.Any("val", val))
	return context.WithValue(ctx, ctxman.key, val)
}

func (ctxman *imapCtxManager) get(ctx context.Context) (*client.Client, error) {
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
