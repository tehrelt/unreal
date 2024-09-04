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

func (s *MailService) Mailboxes(ctx context.Context) ([]*entity.Mailbox, error) {

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

	mbx := make([]*entity.Mailbox, 0, 10)
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		slog.Debug("mailbox", slog.Any("mailbox", m))

		_, err := c.Select(m.Name, false)
		if err != nil {
			return nil, fmt.Errorf("failed to select mailbox %q: %v", m.Name, err)
		}

		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{"\\Seen"}
		ids, err := c.Search(criteria)
		if err != nil {
			return nil, fmt.Errorf("failed to search mailbox %q: %v", m.Name, err)
		}

		mb := &entity.Mailbox{
			Name:        m.Name,
			Attributes:  m.Attributes,
			UnreadCount: len(ids),
		}

		mbx = append(mbx, mb)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to list mailboxes: %v", err)
	}

	return mbx, nil
}
