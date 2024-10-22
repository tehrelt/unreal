package users

import (
	"log/slog"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

type Repository struct {
	pool *pgxpool.Pool
	l    *slog.Logger
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
		l:    slog.With(sl.Module("pg.UserRepository")),
	}
}
