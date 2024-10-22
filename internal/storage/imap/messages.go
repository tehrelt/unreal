package imap

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (r *Repository) Messages(ctx context.Context, in *dto.FetchMessagesDto) (*dto.FetchedMessagesDto, error) {

	fn := "mail.Messages"
	log := r.logger.With(sl.Method(fn))

	c, err := r.ctxman.get(ctx)
	if err != nil {
		return nil, err
	}

	mbox, err := c.Select(in.Mailbox, false)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", fn, err)
	}
	log.Debug("mailbox", slog.Any("mailbox", mbox))
	if mbox.Messages == 0 {
		return &dto.FetchedMessagesDto{
			Messages:    make([]entity.Message, 0),
			Total:       0,
			HasNextPage: false,
		}, nil
	}

	page := int(math.Max(0, float64(in.Page-1)))
	limit := in.Limit
	offset := int(mbox.Messages) - (page * limit)

	cursor := struct {
		Start int
		End   int
	}{
		Start: offset,
		End: func() int {
			expected := offset - limit
			if expected <= 0 {
				return 1
			}
			return expected
		}(),
	}

	toFetch := cursor.Start - cursor.End + 1
	out := &dto.FetchedMessagesDto{
		Messages:    make([]entity.Message, toFetch),
		Total:       int(mbox.Messages),
		HasNextPage: cursor.End != 1,
	}

	log.Debug(
		"fetching messages",
		slog.Any("cursor", cursor),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
		slog.Int("total", int(mbox.Messages)),
		slog.String("range", fmt.Sprintf("AddRange(%d, %d)", cursor.Start, cursor.End)),
	)

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(uint32(cursor.Start), uint32(cursor.End))

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	for m := range messages {

		var msg entity.Message

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

		log.Debug("fetched message", slog.Any("message", msg))

		out.Messages[toFetch-1] = msg
		toFetch--
	}

	// slices.Reverse(mm)

	if err := <-done; err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return out, nil
}
