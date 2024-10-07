//go:build wireinject
// +build wireinject

package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/lib/aes"
	"github.com/tehrelt/unreal/internal/services/authservice"
	"github.com/tehrelt/unreal/internal/services/mailservice"
	mredis "github.com/tehrelt/unreal/internal/storage/redis"
)

func New() (*App, func(), error) {
	panic(wire.Build(
		newApp,
		config.New,
		_redis,

		wire.NewSet(
			mredis.NewSessionStorage,
			wire.Bind(new(authservice.SessionStorage), new(*mredis.SessionStorage)),

			_secretkeyaes,
			aes.NewAesEncryptor,
			wire.Bind(new(authservice.Encryptor), new(*aes.AesEncryptor)),

			authservice.New,
		),

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

func _secretkeyaes(cfg *config.Config) string {
	return cfg.AES.Secret
}
