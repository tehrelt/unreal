package mailservice

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"gopkg.in/gomail.v2"
)

func (s *MailService) Send(ctx context.Context, req *dto.SendMessageDto) error {
	log := slog.With(slog.String("Method", "Mail"))

	u, ok := ctx.Value("user").(*entity.Claims)
	if !ok {
		slog.Warn("no user in context", slog.Any("user", ctx.Value("user")))
		return fmt.Errorf("no user in context")
	}

	log.Info("Sending email")

	dialer := gomail.NewDialer(u.Smtp.Host, u.Smtp.Port, u.Email, u.Password)

	m := gomail.NewMessage()

	// From
	log.Debug("setting from", slog.String("from", u.Email))
	m.SetHeader("From", u.Email)

	// To
	log.Debug("setting to", slog.Any("to", req.To))
	m.SetHeader("To", req.To...)

	// Subject
	log.Debug("setting subject", slog.String("subject", req.Subject))
	m.SetHeader("Subject", req.Subject)

	// Body
	builder := new(strings.Builder)
	if _, err := io.Copy(builder, req.Body); err != nil {
		log.Error("cannot copy body to buffer", sl.Err(err), slog.Any("body", req.Body))
		return fmt.Errorf("cannot copy req.Body: %w", err)
	}
	log.Debug("setting body", slog.Int("len", builder.Len()))
	m.SetBody("text/html", builder.String())

	// Attachments
	for _, a := range req.Attachments {
		log.Debug("attaching file", slog.String("filename", a.Filename))
		m.Attach(a.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
			file, err := a.Open()
			if err != nil {
				log.Error("cannot open attachment", sl.Err(err), slog.Any("filename", a.Filename))
				return err
			}

			if _, err := io.Copy(w, file); err != nil {
				log.Error("cannot copy body to buffer", sl.Err(err), slog.Any("body", file))
			}

			return nil
		}))
	}

	log.Debug("sending message", slog.Any("message", m))
	if err := dialer.DialAndSend(m); err != nil {
		log.Error("cannot send message", sl.Err(err), slog.Any("message", m))
		return fmt.Errorf("cannot send: %w", err)
	}

	return nil
}
