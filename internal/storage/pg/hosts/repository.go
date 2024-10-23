package hosts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/models"
	"github.com/tehrelt/unreal/internal/storage/pg"
)

type Repository struct {
	pool *pgxpool.Pool
	l    *slog.Logger
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
		l:    slog.With(sl.Module("pg.hosts")),
	}
}

func (r *Repository) Find(ctx context.Context, host string) (string, error) {

	fn := "hosts.Find"
	log := r.l.With(sl.Method(fn))

	log.Debug("find host")

	sql, args, err := sq.
		Select("h.picture").
		From(fmt.Sprintf("%s h", pg.HostsTable)).
		Where(sq.Eq{"h.host": host}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Warn("failed to build query")
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}
	defer connection.Release()

	qlog := log.With(slog.String("query", sql), slog.Any("args", args))

	qlog.Debug("querying user")

	var picture string
	if err := connection.QueryRow(ctx, sql, args...).Scan(&picture); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Debug("host not found", slog.String("host", host))
			return "", fmt.Errorf("%s: %w", fn, storage.ErrUserNotFound)
		}
		var pgerr *pgconn.PgError
		if ok := errors.As(err, &pgerr); ok {
			log.Error("unexpected pg error", slog.String("message", pgerr.Message), slog.String("code", pgerr.Code))
		} else {
			log.Error("unexpected error", sl.Err(err))
		}
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	return picture, nil
}

func (r *Repository) Save(ctx context.Context, in *models.CreateHost) error {
	fn := "users.Save"
	log := r.l.With(sl.Method(fn))

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer connection.Release()

	sql, args, err := sq.
		Insert(pg.HostsTable).
		Columns("host", "picture").
		Values(in.Host, in.Picture).
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
			case "23505":
				return fmt.Errorf("%s: %w", fn, storage.ErrHostAlreadyExists)
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
