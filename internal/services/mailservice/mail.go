package mailservice

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/tehrelt/unreal/internal/entity"
	imaps "github.com/tehrelt/unreal/internal/lib/imap"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (s *MailService) Mail(ctx context.Context, mailbox string, num uint32) (*entity.MessageWithBody, error) {

	log := slog.With(slog.String("Method", "Mail"))

	u, ok := ctx.Value("user").(*entity.Claims)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}

	log.Debug("dialing imap", slog.Any("user", u))
	c, cleanup, err := imaps.Dial(u.Email, u.Password, u.Imap.Host, u.Imap.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	defer cleanup()

	mbox, err := c.Select(mailbox, false)
	if err != nil {
		return nil, fmt.Errorf("failed to select mailbox %q: %v", mailbox, err)
	}

	log.Debug("mailbox", slog.Any("mailbox", mbox))

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(num)

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822}

	msg := new(entity.MessageWithBody)
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	m := <-messages
	if m == nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}

	log.Debug("envelope", slog.Any("envelope", m.Envelope), slog.Any("uid", m.Uid))
	log.Debug("message", slog.Any("message", m))

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

		for _, to := range m.Envelope.To {
			msg.To = append(msg.To, entity.AddressInfo{
				Name:    to.PersonalName,
				Address: to.Address(),
			})
		}

		msg.SentDate = m.Envelope.Date
	}

	if r := m.GetBody(&imap.BodySectionName{}); r != nil {
		mr, err := mail.CreateReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to create reader: %v", err)
		}

		for {
			part, err := mr.NextPart()
			if message.IsUnknownEncoding(err) {
				log.Debug("unknown encoding", sl.Err(err))
				continue
			} else if err != nil {
				log.Warn("failed to read part", sl.Err(err))
				break
			}

			log.Debug("part header", slog.Any("header", part.Header))

			switch h := part.Header.(type) {
			case *mail.InlineHeader:
				body, _ := io.ReadAll(part.Body)

				ct, _, err := h.ContentType()
				if err != nil {
					return nil, fmt.Errorf("failed to read content type: %v", err)
				}
				log.Debug("part", slog.String("content-type", ct))

				var bd entity.IBody

				log.Debug("body", slog.String("content-type", ct))

				if strings.Compare(ct, "text/plain") == 0 || strings.Compare(ct, "text/html") == 0 {
					bd = entity.PlainBody(string(body))
				} else {
					bd = entity.BytesBody(body)
				}

				msg.Content = append(msg.Content, entity.Body{
					ContentType: ct,
					Body:        bd,
				})

			case *mail.AttachmentHeader:

				slog.Debug("read attachment", slog.Any("attachment", h))

				// filename, _ := h.Filename()

				// file, err := os.Create(filename)
				// if err != nil {
				// 	return nil, fmt.Errorf("failed to create file: %v", err)
				// }
				// defer file.Close()

				// _, err = io.Copy(file, part.Body)
				// if err != nil {
				// 	return nil, fmt.Errorf("failed to save attachment: %v", err)
				// }
				// slog.Debug("saved attachment", slog.String("filename", filename))
			}
		}
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %v", err)
	}

	return msg, nil
}
