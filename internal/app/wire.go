//go:build wireinject
// +build wireinject

package app

import (
	"context"
	dsa "crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/wire"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/lib/aes"
	dsasigner "github.com/tehrelt/unreal/internal/lib/dsa"
	rsacipher "github.com/tehrelt/unreal/internal/lib/rsa"
	"github.com/tehrelt/unreal/internal/services/authservice"
	"github.com/tehrelt/unreal/internal/services/hostservice"
	"github.com/tehrelt/unreal/internal/services/mailservice"
	"github.com/tehrelt/unreal/internal/storage/fs"
	"github.com/tehrelt/unreal/internal/storage/mail/imap"
	"github.com/tehrelt/unreal/internal/storage/mail/manager"
	"github.com/tehrelt/unreal/internal/storage/mail/smtp"
	"github.com/tehrelt/unreal/internal/storage/pg/hosts"
	usersrepository "github.com/tehrelt/unreal/internal/storage/pg/users"
	"github.com/tehrelt/unreal/internal/storage/pg/vault"
	mredis "github.com/tehrelt/unreal/internal/storage/redis"
)

func New() (*App, func(), error) {
	panic(wire.Build(
		newApp,
		config.New,
		_redis,
		_pgxpool,

		// redis
		mredis.NewSessionStorage,

		// pg
		usersrepository.New,
		hosts.New,
		vault.New,

		// static
		fs.New,

		// protocols
		imap.NewRepository,
		smtp.NewRepository,

		wire.NewSet(
			_rsaPrivateKey,
			_rsaPublicKey,
			rsacipher.New,
		),

		wire.NewSet(
			_secretkeyaes,
			aes.NewCipher,
			aes.NewStringCipher,
		),

		wire.NewSet(
			_dsaPrivateKey,
			_dsaPublicKey,
			dsasigner.New,
		),

		wire.Bind(new(authservice.UserProvider), new(*usersrepository.Repository)),
		wire.Bind(new(authservice.UserSaver), new(*usersrepository.Repository)),
		wire.Bind(new(authservice.UserUpdater), new(*usersrepository.Repository)),
		wire.Bind(new(authservice.FileProvider), new(*fs.FileStorage)),
		wire.Bind(new(authservice.FileUploader), new(*fs.FileStorage)),
		wire.Bind(new(authservice.SessionStorage), new(*mredis.SessionStorage)),
		wire.Bind(new(authservice.Encryptor), new(*aes.StringCipher)),

		wire.Bind(new(mailservice.UserProvider), new(*usersrepository.Repository)),
		wire.Bind(new(mailservice.Repository), new(*imap.Repository)),
		wire.Bind(new(mailservice.Sender), new(*smtp.Repository)),
		wire.Bind(new(mailservice.KnownHostProvider), new(*hosts.Repository)),
		wire.Bind(new(mailservice.Vault), new(*vault.Repository)),
		wire.Bind(new(mailservice.KeyCipher), new(*rsacipher.Cipher)),
		wire.Bind(new(mailservice.Signer), new(*dsasigner.Signer)),

		wire.Bind(new(hostservice.FileUploader), new(*fs.FileStorage)),
		wire.Bind(new(hostservice.HostSaver), new(*hosts.Repository)),

		manager.NewManager,

		authservice.New,
		mailservice.New,
		hostservice.New,
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

func _secretkeyaes(cfg *config.Config) []byte {
	return []byte(cfg.AES.Secret)
}

func _rsaPrivateKey(cfg *config.Config) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(cfg.Jwt.RSA.Private)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key: block.Type = %s", block.Type)
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pkey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not a private key")
	}

	return pkey, nil
}

func _rsaPublicKey(cfg *config.Config) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(cfg.Jwt.RSA.Public)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPubKey, nil
}

func _dsaPrivateKey(cfg *config.Config) (dsa.PrivateKey, error) {

	data, err := os.ReadFile(cfg.DSA.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key: block.Type = %s", block.Type)
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	pkey, ok := privKey.(dsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not a DSA private key")
	}

	return pkey, nil
}

func _dsaPublicKey(cfg *config.Config) (dsa.PublicKey, error) {
	data, err := os.ReadFile(cfg.DSA.PublicKeyFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pub, ok := pubKey.(dsa.PublicKey)
	if !ok {
		return nil, errors.New("not an DSA public key")
	}

	return pub, nil
}
