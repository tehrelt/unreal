package vault

import (
	"context"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage/models"
	"github.com/tehrelt/unreal/internal/storage/pg"
)

func (r *Repository) AppendFiles(ctx context.Context, in *models.AppendFilesArgs) error {
	fn := "vault.AppendFiles"
	log := r.l.With(sl.Method(fn))

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer connection.Release()

	qb := sq.
		Insert(pg.VaultFilesTable).
		Columns(
			"id",
			"message_id",
			"file_name",
			"content_type",
			"hashsum",
			"key",
		).
		PlaceholderFormat(sq.Dollar)

	for _, f := range in.Files {
		qb = qb.Values(f.FileId, in.MessageId, f.Filename, f.ContentType, f.Hashsum, f.Key)
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	log.Debug("executing", slog.String("sql", sql), slog.Any("args", args))

	if _, err := connection.Exec(ctx, sql, args...); err != nil {
		if err != nil {
			log.Error("failed to insert files into pg", sl.Err(err))
			return fmt.Errorf("%s: %w", fn, err)
		}
	}

	return nil
}

func (r *Repository) FileById(ctx context.Context, id string) (*models.VaultFile, error) {

	fn := "vault.AppendFiles"
	log := r.l.With(sl.Method(fn))

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer connection.Release()

	qb := sq.
		Select("id", "message_id", "file_name", "content_type", "hashsum", "key").
		From(pg.VaultFilesTable).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	log.Debug("querying", slog.String("sql", sql), slog.Any("args", args))

	out := new(models.VaultFile)
	if err := connection.QueryRow(ctx, sql, args...).Scan(&out.FileId, &out.MessageId, &out.Filename, &out.ContentType, &out.Hashsum, &out.Key); err != nil {
		log.Error("failed to query file", sl.Err(err))
		return nil, err
	}

	return out, nil
}

func (r *Repository) File(ctx context.Context, messageid, filename string) (*models.VaultFile, error) {

	fn := "vault.AppendFiles"
	log := r.l.With(sl.Method(fn))

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer connection.Release()

	qb := sq.
		Select("id", "message_id", "file_name", "content_type", "hashsum", "key").
		From(pg.VaultFilesTable).
		Where(sq.And{
			sq.Eq{"file_name": filename},
			sq.Eq{"message_id": messageid},
		}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	log.Debug("querying", slog.String("sql", sql), slog.Any("args", args))

	out := new(models.VaultFile)
	if err := connection.QueryRow(ctx, sql, args...).Scan(&out.FileId, &out.MessageId, &out.Filename, &out.ContentType, &out.Hashsum, &out.Key); err != nil {
		log.Error("failed to query file", sl.Err(err))
		return nil, err
	}

	return out, nil
}
