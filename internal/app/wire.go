//go:build wireinject
// +build wireinject

package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/wire"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/lib/aes"
	"github.com/tehrelt/unreal/internal/services/authservice"
	"github.com/tehrelt/unreal/internal/services/mailservice"
	"github.com/tehrelt/unreal/internal/storage/mail/imap"
	"github.com/tehrelt/unreal/internal/storage/mail/manager"
	"github.com/tehrelt/unreal/internal/storage/mail/smtp"
	usersrepository "github.com/tehrelt/unreal/internal/storage/pg/users"
	mredis "github.com/tehrelt/unreal/internal/storage/redis"
)

func New() (*App, func(), error) {
	panic(wire.Build(
		newApp,
		config.New,
		_redis,
		_pgxpool,

		mredis.NewSessionStorage,
		usersrepository.New,

		wire.Bind(new(authservice.UserProvider), new(*usersrepository.Repository)),
		wire.Bind(new(authservice.UserSaver), new(*usersrepository.Repository)),
		wire.Bind(new(authservice.UserUpdater), new(*usersrepository.Repository)),

		wire.Bind(new(authservice.SessionStorage), new(*mredis.SessionStorage)),

		_secretkeyaes,
		aes.NewAesEncryptor,
		wire.Bind(new(authservice.Encryptor), new(*aes.AesEncryptor)),

		authservice.New,

		wire.Bind(new(mailservice.UserProvider), new(*usersrepository.Repository)),

		imap.NewRepository,
		wire.Bind(new(mailservice.Repository), new(*imap.Repository)),

		smtp.NewRepository,
		wire.Bind(new(mailservice.Sender), new(*smtp.Repository)),

		manager.NewManager,

		mailservice.New,
	))
}

func _redis(cfg *config.Config) (*redis.Client, func(), error) {

	conf := cfg.Redis

	ctx := context.Background()

	slog.Debug("connecting to redis", slog.Any("config", conf))
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Password: conf.Pass,
		DB:       conf.DB,
	})

	slog.Debug("ping redis")
	start := time.Now()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, nil, err
	}
	slog.Debug("pinged redis", slog.Duration("took", time.Since(start)))

	return client, func() {
		client.Close()
	}, nil
}

func _pgxpool(cfg *config.Config) (*pgxpool.Pool, func(), error) {

	ctx := context.Background()
	cs := cfg.Pg.ConnectionString()
	db, err := pgxpool.Connect(ctx, cs)
	if err != nil {
		return nil, nil, err
	}

	slog.Debug("connecting to database", slog.String("cs", cs))
	t := time.Now()
	if err := db.Ping(ctx); err != nil {
		slog.Error("failed to connect to database", slog.String("err", err.Error()), slog.String("conn", cs))
		return nil, func() { db.Close() }, err
	}
	slog.Info("connected to database", slog.String("ping", fmt.Sprintf("%2.fs", time.Since(t).Seconds())))

	return db, func() { db.Close() }, nil
}

func _secretkeyaes(cfg *config.Config) string {
	return cfg.AES.Secret
}
