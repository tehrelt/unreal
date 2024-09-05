package mailservice

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/entity"
	imaps "github.com/tehrelt/unreal/internal/lib/imap"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

type MailService struct {
	cfg *config.Config
}

func New(cfg *config.Config) *MailService {
	return &MailService{cfg: cfg}
}

func (s *MailService) Mailboxes(ctx context.Context) ([]*entity.Mailbox, error) {

	log := slog.With(slog.String("Method", "Mailboxes"))

	u, ok := ctx.Value("user").(*entity.Claims)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}

	log.Debug("dialing imap", slog.Any("user", u))
	c, cleanup, err := imaps.Dial(u.Email, u.Password, u.Host, u.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	defer cleanup()

	mbx := make([]*entity.Mailbox, 0, 10)
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		slog.Debug("mailbox", slog.Any("mailbox", m))

		_, err := c.Select(m.Name, false)
		if err != nil {
			return nil, fmt.Errorf("failed to select mailbox %q: %v", m.Name, err)
		}

		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{"\\Seen"}
		ids, err := c.Search(criteria)
		if err != nil {
			return nil, fmt.Errorf("failed to search mailbox %q: %v", m.Name, err)
		}

		mb := &entity.Mailbox{
			Name:        m.Name,
			Attributes:  m.Attributes,
			UnreadCount: len(ids),
		}

		mbx = append(mbx, mb)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to list mailboxes: %v", err)
	}

	return mbx, nil
}

func (s *MailService) Messages(ctx context.Context, mailbox string) ([]*entity.Message, error) {
	log := slog.With(slog.String("Method", "Messages"))

	u, ok := ctx.Value("user").(*entity.Claims)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}

	log.Debug("dialing imap", slog.Any("user", u))
	c, cleanup, err := imaps.Dial(u.Email, u.Password, u.Host, u.Port)
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
	seqSet.AddRange(1, mbox.Messages)

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}

	mm := make([]*entity.Message, 0, mbox.Messages)

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
			msg.From = entity.From{
				Name:    from.PersonalName,
				Address: from.Address(),
			}
			msg.SentDate = m.Envelope.Date.String()
		}

		slog.Debug("fetched message", slog.Any("message", msg))

		mm = append(mm, msg)
	}

	slices.Reverse(mm)

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %v", err)
	}

	return mm, nil
}

func (s *MailService) Mail(ctx context.Context, mailbox string, num uint32) (*entity.Message, error) {

	log := slog.With(slog.String("Method", "Mail"))

	u, ok := ctx.Value("user").(*entity.Claims)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}

	log.Debug("dialing imap", slog.Any("user", u))
	c, cleanup, err := imaps.Dial(u.Email, u.Password, u.Host, u.Port)
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

	msg := new(entity.Message)
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	m := <-messages
	if m == nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}

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
		msg.From = entity.From{
			Name:    from.PersonalName,
			Address: from.Address(),
		}
		msg.SentDate = m.Envelope.Date.String()
	}

	if r := m.GetBody(&imap.BodySectionName{}); r != nil {
		mr, err := mail.CreateReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to create reader: %v", err)
		}

		for {
			part, err := mr.NextPart()
			if message.IsUnknownEncoding(err) {
				slog.Debug("unknown encoding", sl.Err(err))
				continue
			} else if err != nil {
				slog.Warn("failed to read part", sl.Err(err))
				break
			}

			switch h := part.Header.(type) {
			case *mail.InlineHeader:
				body, _ := io.ReadAll(part.Body)
				slog.Debug("read body", slog.String("body", string(body)))
				msg.Body = string(body)
			case *mail.AttachmentHeader:
				filename, _ := h.Filename()

				file, err := os.Create(filename)
				if err != nil {
					return nil, fmt.Errorf("failed to create file: %v", err)
				}
				defer file.Close()

				_, err = io.Copy(file, part.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to save attachment: %v", err)
				}
				slog.Debug("saved attachment", slog.String("filename", filename))
			}
		}
	}

	slog.Debug("fetched message", slog.Any("message", msg))

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %v", err)
	}

	return msg, nil
}
