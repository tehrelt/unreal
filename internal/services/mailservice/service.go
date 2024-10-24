package mailservice

import (
	"context"
	"io"
	"log/slog"

	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/models"
)

type Repository interface {
	Mailboxes(ctx context.Context) ([]*entity.Mailbox, error)
	Messages(ctx context.Context, in *dto.FetchMessagesDto) (*dto.FetchedMessagesDto, error)
	Message(ctx context.Context, mailbox string, mailnum uint32) (*entity.MessageWithBody, error)
	SaveMessageToFolderByAttribute(ctx context.Context, attr string, msg io.Reader) error
	Attachment(ctx context.Context, mailbox string, mailnum uint32, target string) (out io.Reader, ct string, err error)
}

type Sender interface {
	Send(ctx context.Context, req *dto.SendMessageDto) (io.Reader, error)
}

type UserProvider interface {
	Find(ctx context.Context, email string) (*models.User, error)
}

type KnownHostProvider interface {
	Find(ctx context.Context, host string) (string, error)
}

type Vault interface {
	Insert(ctx context.Context, in *models.VaultRecord) error
	Find(ctx context.Context, messageId string) (*models.VaultRecord, error)
}

type KeyCipher interface {
	Encrypt(io.Reader) (io.Reader, error)
	Decrypt(io.Reader) (io.Reader, error)
}

type Service struct {
	cfg          *config.Config
	m            storage.Manager
	r            Repository
	l            *slog.Logger
	sender       Sender
	userProvider UserProvider
	hostProvider KnownHostProvider
	vault        Vault
	keyCipher    KeyCipher
}

func New(
	cfg *config.Config,
	manager storage.Manager,
	r Repository,
	sender Sender,
	userProvider UserProvider,
	hostProvider KnownHostProvider,
	vault Vault,
	cipher KeyCipher,
) *Service {
	return &Service{
		cfg:          cfg,
		m:            manager,
		r:            r,
		l:            slog.With(sl.Method("mailservice.MailService")),
		sender:       sender,
		userProvider: userProvider,
		hostProvider: hostProvider,
		vault:        vault,
		keyCipher:    cipher,
	}
}
