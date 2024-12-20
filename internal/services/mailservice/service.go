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
	Message(ctx context.Context, mailbox string, mailnum uint32) (*models.Message, error)
	Raw(ctx context.Context, mailbox string, mailnum uint32) (io.Reader, error)
	Delete(ctx context.Context, mailbox string, mailnum uint32) error
	Attachment(ctx context.Context, mailbox string, mailnum uint32, target string) (out *models.Attachment, err error)
	IsMessageEncrypted(ctx context.Context, mailbox string, num uint32) (vaultId string, err error)
	Health(ctx context.Context) (bool, error)

	SaveSentMessage(ctx context.Context, msg io.Reader) error
	SaveDraftMessage(ctx context.Context, msg io.Reader) error
}

type Sender interface {
	Literal(ctx context.Context, req *models.SendMessage) (io.Reader, error)
	Send(ctx context.Context, req *models.SendMessage) (io.Reader, error)
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
	File(ctx context.Context, messageId, name string) (*models.VaultFile, error)
	FileById(ctx context.Context, id string) (*models.VaultFile, error)
	AppendFiles(ctx context.Context, in *models.AppendFilesArgs) error
}

type Signer interface {
	Sign(data []byte) []byte
	Verify(data, signature []byte) error
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
	signer       Signer
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
	signer Signer,
) *Service {
	return &Service{
		cfg:          cfg,
		m:            manager,
		r:            r,
		sender:       sender,
		userProvider: userProvider,
		hostProvider: hostProvider,
		vault:        vault,
		keyCipher:    cipher,
		signer:       signer,
		l:            slog.With(sl.Method("mailservice.MailService")),
	}
}
