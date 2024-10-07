package mailservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/entity"
	imaps "github.com/tehrelt/unreal/internal/lib/imap"
)

func (s *MailService) Mailboxes(ctx context.Context) ([]*entity.Mailbox, error) {

	log := slog.With(slog.String("Method", "Mailboxes"))

	u, ok := ctx.Value("user").(*entity.SessionInfo)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}

	log.Debug("dialing imap", slog.Any("user", u))
	c, cleanup, err := imaps.Dial(u.Email, u.Password, u.Imap.Host, u.Imap.Port)
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

		unread, err := s.unreadMessage(ctx, c)
		if err != nil {
			return nil, err
		}

		mb := &entity.Mailbox{
			Name:        entity.NewMailboxName(m.Name),
			Attributes:  m.Attributes,
			UnreadCount: unread,
		}

		mbx = append(mbx, mb)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to list mailboxes: %v", err)
	}

	return mbx, nil
}
