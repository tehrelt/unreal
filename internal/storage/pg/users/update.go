package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage/models"
	"github.com/tehrelt/unreal/internal/storage/pg"
)

func (r *Repository) Update(ctx context.Context, in *models.UpdateUser) (err error) {
	fn := "users.Update"
	log := r.l.With(sl.Method(fn))

	if in.Name == nil && in.ProfilePicture == nil {
		log.Warn("nothing to update", slog.Any("in", in))
		return nil
	}

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer connection.Release()

	tx, err := connection.Begin(ctx)
	if err != nil {
		log.Error("failed to begin transaction", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			log.Error("failed to commit transaction", sl.Err(commitErr))
			err = fmt.Errorf("%s: %w", fn, commitErr)
		}
	}()

	qb := sq.
		Update(pg.UserTable).
		Where(sq.Eq{"email": in.Email}).
		PlaceholderFormat(sq.Dollar)

	if in.Name != nil {
		qb = qb.Set("name", in.Name)
	}

	if in.ProfilePicture != nil {
		if err := r.setProfilePicture(ctx, tx, in.Email, *in.ProfilePicture); err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}
	}

	sql, args, err := qb.
		ToSql()
	if err != nil {
		log.Error("failed to build query", sl.Err(err))
	}

	qlog := log.With(slog.String("query", sql), slog.Any("args", args))

	qlog.Debug("executing query")
	if _, err := tx.Exec(ctx, sql, args...); err != nil {
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

func (r *Repository) isProfilePictureSet(ctx context.Context, tx pgx.Tx, email string) (bool, error) {
	fn := "users.isProfilePictureSet"
	log := r.l.With(sl.Method(fn))

	sql, args, err := sq.
		Select("profile_picture").
		From(pg.ProfilePictureTable).
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Error("failed to build query", sl.Err(err))
		return false, fmt.Errorf("%s: %w", fn, err)
	}

	qlog := log.With(slog.String("query", sql), slog.Any("args", args))
	qlog.Debug("executing query")

	var pfp string
	if err := tx.QueryRow(ctx, sql, args...).Scan(&pfp); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		log.Error("unexpected error", sl.Err(err))
		return false, fmt.Errorf("%s: %w", fn, err)
	}

	return pfp != "", nil
}

func (r *Repository) setProfilePicture(ctx context.Context, tx pgx.Tx, email, pfp string) error {
	fn := "users.updateProfilePicture"
	log := r.l.With(sl.Method(fn))

	var sql string
	var args []any
	var err error

	set, err := r.isProfilePictureSet(ctx, tx, email)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	if set {
		log.Debug("updating pfp", slog.String("email", email))
		sql, args, err = sq.
			Update(pg.ProfilePictureTable).
			Set("profile_picture", pfp).
			Where(sq.Eq{"email": email}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
	} else {
		log.Debug("inserting pfp", slog.String("email", email))
		sql, args, err = sq.
			Insert(pg.ProfilePictureTable).
			Columns("email", "profile_picture").
			Values(email, pfp).
			PlaceholderFormat(sq.Dollar).
			ToSql()
	}
	if err != nil {
		log.Error("failed to build query", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	qlog := log.With(slog.String("query", sql), slog.Any("args", args))
	qlog.Debug("executing query")
	if _, err := tx.Exec(ctx, sql, args...); err != nil {
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
