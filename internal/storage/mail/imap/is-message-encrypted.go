package imap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
)

func (r *Repository) IsMessageEncrypted(ctx context.Context, mailbox string, num uint32) (vaultId string, err error) {

	fn := "mail.Message"
	log := slog.With(sl.Method(fn))

	c, err := r.ctxman.get(ctx)
	if err != nil {
		return "", err
	}

	mbox, err := c.Select(mailbox, false)
	if err != nil {
		return "", fmt.Errorf("failed to select mailbox %q: %v", mailbox, err)
	}

	log.Debug("mailbox", slog.Any("mailbox", mbox))

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(num)

	items := []imap.FetchItem{imap.FetchRFC822, imap.FetchRFC822Header}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	m := <-messages
	if m == nil {
		return "", fmt.Errorf("failed to fetch message: %w", err)
	}

	if r := m.GetBody(&imap.BodySectionName{}); r != nil {
		mr, err := mail.CreateReader(r)
		if err != nil {
			return "", fmt.Errorf("failed to create reader: %v", err)
		}

		vaultId = mr.Header.Get(storage.EncryptionHeader)
	}

	if err := <-done; err != nil {
		return "", fmt.Errorf("failed to fetch messages: %v", err)
	}

	return
}
