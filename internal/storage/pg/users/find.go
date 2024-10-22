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
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/models"
	"github.com/tehrelt/unreal/internal/storage/pg"
)

func (r *Repository) Find(ctx context.Context, email string) (*models.User, error) {

	fn := "users.Find"
	log := r.l.With(sl.Method(fn))

	log.Debug("find user by email", slog.String("email", email))

	sql, args, err := sq.
		Select("u.id, u.email, u.name, u.created_at, u.updated_at, pfp.profile_picture").
		From(fmt.Sprintf("%s u", pg.UserTable)).
		LeftJoin(fmt.Sprintf("%s pfp on u.id = pfp.user_id", pg.ProfilePictureTable)).
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Warn("failed to build query")
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer connection.Release()

	qlog := log.With(slog.String("query", sql), slog.Any("args", args))

	qlog.Debug("querying user")

	var u models.User
	if err := connection.QueryRow(ctx, sql, args...).Scan(&u.Id, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt, &u.ProfilePicture); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Debug("user not found", slog.String("email", email))
			return nil, fmt.Errorf("%s: %w", fn, storage.ErrUserNotFound)
		}
		var pgerr *pgconn.PgError
		if ok := errors.As(err, &pgerr); ok {
			log.Error("unexpected pg error", slog.String("message", pgerr.Message), slog.String("code", pgerr.Code))
		} else {
			log.Error("unexpected error", sl.Err(err))
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &u, nil
}
