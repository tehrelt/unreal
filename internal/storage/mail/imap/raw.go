package imap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage/models"
)

func (r *Repository) Raw(ctx context.Context, mailbox string, num uint32) (io.Reader, error) {

	fn := "mail.Message"
	log := slog.With(sl.Method(fn))

	c, err := r.ctxman.get(ctx)
	if err != nil {
		return nil, err
	}

	mbox, err := c.Select(mailbox, false)
	if err != nil {
		return nil, fmt.Errorf("failed to select mailbox %q: %v", mailbox, err)
	}

	log.Debug("mailbox", slog.Any("mailbox", mbox))

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(num)

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822, imap.FetchRFC822Header}

	msg := new(models.Message)
	html := new(bytes.Buffer)
	msg.Body = html
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	m := <-messages
	if m == nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}

	body := m.GetBody(&imap.BodySectionName{})
	if r == nil {
		log.Error("failed to get message body", slog.Any("message", m))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %v", err)
	}

	return body, nil
}
