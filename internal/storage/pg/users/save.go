package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage/models"
	"github.com/tehrelt/unreal/internal/storage/pg"
)

func (r *Repository) Save(ctx context.Context, in *models.CreateUser) error {
	fn := "users.Save"
	log := r.l.With(sl.Method(fn))

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer connection.Release()

	sql, args, err := sq.
		Insert(pg.UserTable).
		Columns("id", "email").
		Values(in.Id, in.Email).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Error("failed to build query", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	qlog := log.With(slog.String("query", sql), slog.Any("args", args))

	qlog.Debug("executing")
	if _, err := connection.Exec(ctx, sql, args...); err != nil {
		var pgerr *pgconn.PgError
		if ok := errors.As(err, &pgerr); ok {
			switch pgerr.Code {
			default:
				log.Error("unexpected pg error", slog.String("message", pgerr.Message), slog.String("code", pgerr.Code))
				return fmt.Errorf("%s: %w", fn, err)
			}
		} else {
			log.Error("unexpected error", sl.Err(err))
			return fmt.Errorf("%s: %w", fn, err)
		}
	}

	return nil
}
