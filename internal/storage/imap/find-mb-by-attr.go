package imap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/domain"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (r *Repository) findFolderByAttr(ctx context.Context, attribute string) (string, error) {

	fn := "imap.findFolderByAttr"
	log := slog.With(sl.Method(fn))

	c, err := r.ctxman.get(ctx)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	var folder string

	for m := range mailboxes {
		slog.Debug(
			"comparing mailbox to attribute",
			slog.String("mailbox", m.Name),
			slog.String("attribute", attribute),
		)
		for _, attr := range m.Attributes {
			if attr == attribute {
				folder = m.Name
				log.Debug("found mailbox", slog.String("folder", folder))

				return folder, nil
			}
		}
	}

	if err := <-done; err != nil {
		return "", err
	}

	if folder == "" {
		return "", fmt.Errorf("%s: %w", fn, domain.ErrMailboxNotFound)
	}

	return folder, nil
}
