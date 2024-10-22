package smtp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"gopkg.in/gomail.v2"
)

func (r *Repository) Send(ctx context.Context, req *dto.SendMessageDto) (io.Reader, error) {
	fn := "smtp.Send"
	log := r.l.With(sl.Method(fn))

	u, err := r.ctxman.session(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	dialer, err := r.dial(u)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	m := gomail.NewMessage()

	log.Debug("setting from", slog.String("from", req.From.String()))
	m.SetHeader("From", req.From.String())

	// To
	log.Debug("setting to", slog.Any("to", req.To))
	m.SetHeader("To", lo.Map(req.To, func(r *dto.MailRecord, _ int) string { return r.String() })...)

	// Subject
	log.Debug("setting subject", slog.String("subject", req.Subject))
	m.SetHeader("Subject", req.Subject)

	// Body
	builder := new(strings.Builder)
	if _, err := io.Copy(builder, req.Body); err != nil {
		log.Error("cannot copy body to buffer", sl.Err(err), slog.Any("body", req.Body))
		return nil, fmt.Errorf("cannot copy req.Body: %w", err)
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

	buf := new(bytes.Buffer)
	if _, err := m.WriteTo(buf); err != nil {
		log.Error("cannot write message to buffer", sl.Err(err), slog.Any("message", m))
		return nil, fmt.Errorf("cannot write to buf: %w", err)
	}

	log.Debug("sending message", slog.Any("message", m))
	if err := dialer.DialAndSend(m); err != nil {
		log.Error("cannot send message", sl.Err(err), slog.Any("message", m))
		return nil, fmt.Errorf("cannot send: %w", err)
	}

	return buf, nil
}
