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

type Repository interface {
	Mailboxes(ctx context.Context) ([]*entity.Mailbox, error)
	Messages(ctx context.Context, in *dto.FetchMessagesDto) (*dto.FetchedMessagesDto, error)
	Message(ctx context.Context, mailbox string, mailnum uint32) (*entity.MessageWithBody, error)
}

type Sender interface {
	Send(ctx context.Context, req *dto.SendMessageDto) error
}

type Service struct {
	cfg    *config.Config
	m      storage.Manager
	r      Repository
	l      *slog.Logger
	sender Sender
}

func New(cfg *config.Config, manager storage.Manager, r Repository, sender Sender) *Service {
	return &Service{
		cfg:    cfg,
		m:      manager,
		r:      r,
		l:      slog.With(sl.Method("mailservice.MailService")),
		sender: sender,
	}
}
