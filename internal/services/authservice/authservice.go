package authservice

import (
	"context"
	"log/slog"

	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/entity"
)

type SessionStorage interface {
	Find(ctx context.Context, id string) (*entity.SessionInfo, error)
	Save(ctx context.Context, in *entity.SessionInfo) (id string, err error)
}

type Encryptor interface {
	Encrypt(in string) (string, error)
	Decrypt(in string) (string, error)
}

type AuthService struct {
	cfg       *config.Config
	logger    *slog.Logger
	sessions  SessionStorage
	encryptor Encryptor
}

func New(cfg *config.Config, sessions SessionStorage, encryptor Encryptor) *AuthService {
	return &AuthService{
		cfg:       cfg,
		sessions:  sessions,
		encryptor: encryptor,
		logger:    slog.Default().With(slog.String("struct", "AuthService")),
	}
}
