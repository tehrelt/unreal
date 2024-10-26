package vault

import (
	"context"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage/models"
	"github.com/tehrelt/unreal/internal/storage/pg"
)

// var _ mailservice.Vault = new(Repository)

type Repository struct {
	pool *pgxpool.Pool
	l    *slog.Logger
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool, slog.With(sl.Module("pg.vaultrepository"))}
}

func (r *Repository) Find(ctx context.Context, id string) (*models.VaultRecord, error) {
	fn := "valultrepository.Find"
	log := r.l.With(sl.Method(fn))

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer conn.Release()

	sql, args, err := sq.
		Select("id", "key", "hashsum").
		From(pg.VaultTable).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Error("failed to build query", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	qlog := log.With(slog.String("query", sql), slog.Any("args", args))

	qlog.Debug("executing query")
	var record models.VaultRecord
	if err = conn.QueryRow(ctx, sql, args...).Scan(&record.Id, &record.Key, &record.Hashsum); err != nil {
		log.Error("failed to query", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &record, nil
}

func (r *Repository) Insert(ctx context.Context, in *models.VaultRecord) error {
	fn := "valultrepository.Insert"
	log := r.l.With(sl.Method(fn))

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer conn.Release()

	sql, args, err := sq.
		Insert(pg.VaultTable).
		Columns("id", "key", "hashsum").
		Values(in.Id, in.Key, in.Hashsum).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Error("failed to build query", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	qlog := log.With(slog.String("query", sql), slog.Any("args", args))

	qlog.Debug("executing query")

	if _, err = conn.Exec(ctx, sql, args...); err != nil {
		log.Error("failed to execute query", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
