package mailservice

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
	imaps "github.com/tehrelt/unreal/internal/lib/imap"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (s *MailService) GetAttachment(ctx context.Context, mailbox string, mailnum uint32, target string) (r io.Reader, ct string, err error) {

	log := slog.With(slog.String("Method", "Mail"))

	log.Debug(
		"find an attachment",
		slog.String("mailbox", mailbox),
		slog.Any("mailnum", mailnum),
		slog.String("target", target),
	)

	u, ok := ctx.Value("user").(*entity.Claims)
	if !ok {
		return nil, "", fmt.Errorf("no user in context")
	}

	log.Debug("dialing imap", slog.Any("user", u))
	c, cleanup, err := imaps.Dial(u.Email, u.Password, u.Imap.Host, u.Imap.Port)
	if err != nil {
		return nil, "", fmt.Errorf("failed to connect: %v", err)
	}
	defer cleanup()

	_, err = c.Select(mailbox, false)
	if err != nil {
		return nil, "", fmt.Errorf("failed to select mailbox %q: %v", mailbox, err)
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(mailnum)

	items := []imap.FetchItem{imap.FetchRFC822}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	m := <-messages
	if m == nil {
		return nil, "", fmt.Errorf("failed to fetch message: %w", err)
	}

	if rr := m.GetBody(&imap.BodySectionName{}); rr != nil {
		mr, err := mail.CreateReader(rr)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create reader: %v", err)
		}

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
			case *mail.InlineHeader:
				cid := h.Get("Content-Id")
				log.Debug("get a cid", slog.Any("cid", cid))
				if cid == "" {
					slog.Debug("cid is empty")
					continue
				}

				ct, _, err = h.ContentType()
				if err != nil {
					return nil, "", fmt.Errorf("failed to get content type: %v", err)
				}

				contains := strings.Contains(cid, target)

				log.Debug("comparing cids", slog.Any("ct", cid), slog.Any("targetCid", target), slog.Any("res", contains))

				if contains {
					log.Debug("found a cid")
					buf := new(bytes.Buffer)
					buf.ReadFrom(part.Body)
					r = buf
				}

			case *mail.AttachmentHeader:

				filename, err := h.Filename()
				if err != nil {
					slog.Debug("failed to read filename", sl.Err(err))
					filename = fmt.Sprintf("file-%s", uuid.New())
				}

				if strings.Compare(filename, target) == 0 {
					ct, _, err = h.ContentType()
					if err != nil {
						slog.Warn("failed to get content type", sl.Err(err))
						ct = "application/octet-stream"
					}

					log.Debug("found a filename", slog.Any("filename", filename), slog.Any("targetCid", target))
					buf := new(bytes.Buffer)
					buf.ReadFrom(part.Body)
					r = buf
					break
				}
			}
		}
	}

	if r == nil {
		return nil, "", fmt.Errorf("failed to find attachment")
	}

	return
}
