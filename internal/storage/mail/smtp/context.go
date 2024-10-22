package smtp

import (
	"context"
	"log/slog"

	gctx "github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
)

var defaultManager = newCtxManager()

type smtpCtxManager struct {
	key gctx.CtxKey
	l   *slog.Logger
}

func newCtxManager() *smtpCtxManager {
	return &smtpCtxManager{
		key: dialerKey,
		l: slog.With(
			sl.Module("smtpCtxManager"),
			slog.Any("key", dialerKey),
		),
	}
}

func (man *smtpCtxManager) set(ctx context.Context, session *entity.SessionInfo) context.Context {
	fn := "context.Set"
	log := man.l.With(sl.Method(fn))

	log.Debug("setting context", slog.Any("val", session))
	return context.WithValue(ctx, man.key, session)
}

func (man *smtpCtxManager) session(ctx context.Context) (*entity.SessionInfo, error) {
	fn := "context.session"
	log := man.l.With(sl.Method(fn))

	val := ctx.Value(man.key)
	if val == nil {
		log.Error("empty context")
		return nil, storage.ErrNotEnrichedContext
	}

	c, ok := val.(*entity.SessionInfo)
	if !ok {
		return nil, storage.ErrInvalidValue
	}
	log.Debug("extracted connection from context", slog.Any("c", c))

	return c, nil
}
