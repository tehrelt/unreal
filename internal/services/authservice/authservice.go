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

type FileProvider interface {
	File(ctx context.Context, filename string) (*models.File, error)
}

type FileUploader interface {
	Upload(ctx context.Context, entry *models.File) error
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
	fileProvider FileProvider
	fileUploader FileUploader
}

func New(
	cfg *config.Config,
	sessions SessionStorage,
	encryptor Encryptor,
	userProvider UserProvider,
	userSaver UserSaver,
	userUpdater UserUpdater,
	fileProvider FileProvider,
	fileUploader FileUploader,
) *AuthService {
	return &AuthService{
		cfg:          cfg,
		logger:       slog.Default().With(slog.String("struct", "AuthService")),
		sessions:     sessions,
		encryptor:    encryptor,
		userSaver:    userSaver,
		userUpdater:  userUpdater,
		userProvider: userProvider,
		fileProvider: fileProvider,
		fileUploader: fileUploader,
	}
}
