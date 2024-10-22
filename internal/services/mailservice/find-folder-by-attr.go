package mailservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/tehrelt/unreal/internal/domain"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (ms *Service) findFolderByAttr(_ context.Context, c *client.Client, attribute string) (string, error) {

	fn := "mailservice.findFolderByAttr"
	log := slog.With(sl.Method(fn))

	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	var folder string

	for m := range mailboxes {
		slog.Debug(
			"comapring mailbox to attribute",
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
