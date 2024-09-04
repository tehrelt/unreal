package mailservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/entity"
	imaps "github.com/tehrelt/unreal/internal/lib/imap"
)

type MailService struct {
	cfg *config.Config
}

func New(cfg *config.Config) *MailService {
	return &MailService{cfg: cfg}
}

func (s *MailService) Mailboxes(ctx context.Context) ([]*imap.MailboxInfo, error) {

	log := slog.With(slog.String("Method", "Mailboxes"))

	u, ok := ctx.Value("user").(*entity.Claims)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}

	log.Debug("dialing imap", slog.Any("user", u))
	c, cleanup, err := imaps.Dial(u.Email, u.Password, u.Host, u.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	defer cleanup()

	mbx := make([]*imap.MailboxInfo, 0, 10)

	mailboxes := make(chan *imap.MailboxInfo, 10)

	done := make(chan error, 1)

	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		slog.Debug("mailbox", slog.Any("mailbox", m))
		mbx = append(mbx, m)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to list mailboxes: %v", err)
	}

	slog.Debug("mailboxes", slog.Any("mailboxes", mbx))

	return mbx, nil
}
