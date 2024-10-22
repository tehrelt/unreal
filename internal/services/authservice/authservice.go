package authservice

import (
	"context"
	"log/slog"
	"time"

	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/storage/models"
)

type SessionStorage interface {
	Find(ctx context.Context, id string) (*entity.SessionInfo, error)
	Save(ctx context.Context, in *entity.SessionInfo, ttl ...time.Duration) (id string, err error)
}

type UserProvider interface {
	Find(ctx context.Context, email string) (*models.User, error)
}

type UserSaver interface {
	Save(ctx context.Context, in *models.CreateUser) error
}

type UserUpdater interface {
	Update(ctx context.Context, in *models.UpdateUser) error
}

type Encryptor interface {
	Encrypt(in string) (string, error)
	Decrypt(in string) (string, error)
}

type AuthService struct {
	cfg          *config.Config
	logger       *slog.Logger
	sessions     SessionStorage
	encryptor    Encryptor
	userProvider UserProvider
	userSaver    UserSaver
	userUpdater  UserUpdater
}

func New(cfg *config.Config, sessions SessionStorage, encryptor Encryptor, userProvider UserProvider, userSaver UserSaver, userUpdater UserUpdater) *AuthService {
	return &AuthService{
		cfg:          cfg,
		sessions:     sessions,
		encryptor:    encryptor,
		logger:       slog.Default().With(slog.String("struct", "AuthService")),
		userProvider: userProvider,
		userSaver:    userSaver,
		userUpdater:  userUpdater,
	}
}
