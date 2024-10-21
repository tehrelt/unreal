package mailservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/domain"
	"github.com/tehrelt/unreal/internal/entity"
	imaps "github.com/tehrelt/unreal/internal/lib/imap"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (s *MailService) Mailboxes(ctx context.Context) ([]*entity.Mailbox, error) {

	fn := "mailservice.Mailboxes"
	log := slog.With(sl.Method(fn))

	u, ok := ctx.Value("user").(*entity.SessionInfo)
	if !ok {
		return nil, fmt.Errorf("%s: %w", fn, domain.ErrUserNotInContext)
	}

	log.Debug("dialing imap", slog.Any("user", u))
	c, cleanup, err := imaps.Dial(u.Email, u.Password, u.Imap.Host, u.Imap.Port)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer cleanup()

	mbx := make([]*entity.Mailbox, 0, 10)
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		_, err := c.Select(m.Name, true)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		unread, err := s.unreadMessage(ctx, c)
		if err != nil {
			return nil, err
		}

		slog.Debug("mailbox attributes", slog.String("mailbox", m.Name), slog.Any("attributes", m.Attributes))

		mb := &entity.Mailbox{
			Name:        entity.NewMailboxName(m.Name),
			Attributes:  m.Attributes,
			UnreadCount: unread,
		}

		mbx = append(mbx, mb)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return mbx, nil
}
