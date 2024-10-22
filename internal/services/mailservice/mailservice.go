package mailservice

import (
	"context"
	"log/slog"

	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
)

type MailRepository interface {
	Mailboxes(ctx context.Context) ([]*entity.Mailbox, error)
	Messages(ctx context.Context, in *dto.FetchMessagesDto) (*dto.FetchedMessagesDto, error)
	Message(ctx context.Context, mailbox string, mailnum uint32) (*entity.MessageWithBody, error)
}

type MailService struct {
	cfg *config.Config
	m   storage.Manager
	r   MailRepository
	l   *slog.Logger
}

func New(cfg *config.Config, manager storage.Manager, r MailRepository) *MailService {
	return &MailService{
		cfg: cfg,
		m:   manager,
		r:   r,
		l:   slog.With(sl.Method("mailservice.MailService")),
	}
}
