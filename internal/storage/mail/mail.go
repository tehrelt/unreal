package mail

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emersion/go-imap"
	gctx "github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	mctx "github.com/tehrelt/unreal/internal/storage/mail/context"
)

type MailRepository struct {
	ctxman *mctx.MailContextManager
	logger *slog.Logger
}

func NewMailRepository(key gctx.CtxKey) *MailRepository {
	return &MailRepository{
		ctxman: mctx.New(key),
		logger: slog.With(sl.Module("mail.MailRepository")),
	}
}

func (r *MailRepository) Mailboxes(ctx context.Context) ([]*entity.Mailbox, error) {

	fn := "mail.Mailboxes"
	log := r.logger.With(sl.Method(fn))

	conn, err := r.ctxman.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	mbx := make([]*entity.Mailbox, 0, 10)
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)

	go func() {
		done <- conn.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		log.Debug("fetching mailbox", slog.String("name", m.Name))
		_, err := conn.Select(m.Name, true)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		unread, err := r.unread(ctx)
		if err != nil {
			return nil, err
		}

		mb := &entity.Mailbox{
			Name:        m.Name,
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
