package imap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/google/uuid"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/models"
)

func (r *Repository) Message(ctx context.Context, mailbox string, num uint32) (*models.Message, error) {

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
		msg.SeqNum = m.SeqNum
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

		msg.VaultId = mr.Header.Get(storage.EncryptionHeader)

		for {
			part, err := mr.NextPart()

			if message.IsUnknownEncoding(err) {
				log.Debug("unknown encoding", sl.Err(err))
				continue
			} else if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				log.Warn("failed to read part", sl.Err(err))
				break
			}

			switch h := part.Header.(type) {
			case *mail.Header:
				log.Debug("read header", slog.Any("headers", h))

			case *mail.InlineHeader:

				ct, _, err := h.ContentType()
				if err != nil {
					return nil, fmt.Errorf("failed to read content type: %v", err)
				}

				if strings.Compare(ct, "text/html") == 0 {
					html.ReadFrom(part.Body)
				} else if strings.HasPrefix(ct, "image/") {
					cid := h.Get("Content-Id")
					if cid == "" {
						slog.Debug("cid is empty")
					}

					cid = strings.Trim(cid, "<>")

					msg.Attachments = append(msg.Attachments, entity.Attachment{
						ContentId:   cid,
						Filename:    cid,
						ContentType: ct,
					})
				}

			case *mail.AttachmentHeader:

				filename, err := h.Filename()
				if err != nil {
					slog.Debug("failed to read filename", sl.Err(err))
					filename = fmt.Sprintf("file-%s", uuid.New())
				}

				ct, _, err := h.ContentType()
				if err != nil {
					slog.Debug("failed to read content type", sl.Err(err))
					ct = "application/octet-stream"
				}

				msg.Attachments = append(msg.Attachments, entity.Attachment{
					ContentId:   filename,
					Filename:    filename,
					ContentType: ct,
				})
			default:
				log.Debug("unknown header type", slog.String("type", fmt.Sprintf("%T", h)), slog.Any("header", h))
			}
		}
	}

	slog.Info("fetched message", slog.Any("mail", msg))

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %v", err)
	}

	return msg, nil
}
