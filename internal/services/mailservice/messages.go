package mailservice

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/entity"
	imaps "github.com/tehrelt/unreal/internal/lib/imap"
)

func (s *MailService) Messages(ctx context.Context, mailbox string) ([]*entity.Message, int, error) {
	log := slog.With(slog.String("Method", "Messages"))

	u, ok := ctx.Value("user").(*entity.SessionInfo)
	if !ok {
		return nil, 0, fmt.Errorf("no user in context")
	}

	log.Debug("dialing imap", slog.Any("user", u))
	c, cleanup, err := imaps.Dial(u.Email, u.Password, u.Imap.Host, u.Imap.Port)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to connect: %v", err)
	}
	defer cleanup()

	mbox, err := c.Select(mailbox, false)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to select mailbox %q: %v", mailbox, err)
	}

	log.Debug("mailbox", slog.Any("mailbox", mbox))

	mm := make([]*entity.Message, 0, mbox.Messages)
	if mbox.Messages == 0 {
		return mm, int(mbox.Messages), nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(mbox.Messages, 1)

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	for m := range messages {
		log.Debug("message", slog.Any("message", m))

		msg := new(entity.Message)

		msg.IsRead = false
		for _, flag := range m.Flags {
			if flag == imap.SeenFlag {
				msg.IsRead = true
				break
			}
		}

		if m.Envelope != nil {
			msg.Id = m.SeqNum
			msg.Subject = m.Envelope.Subject
			from := m.Envelope.From[0]
			msg.From = entity.AddressInfo{
				Name:    from.PersonalName,
				Address: from.Address(),
			}
			msg.SentDate = m.Envelope.Date
		}

		slog.Debug("fetched message", slog.Any("message", msg))

		mm = append(mm, msg)
	}

	slices.Reverse(mm)

	if err := <-done; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch messages: %v", err)
	}

	return mm, int(mbox.Messages), nil
}
