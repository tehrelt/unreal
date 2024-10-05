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
		return fmt.Errorf("no user in context")
	}

	log.Info("Sending email")

	dialer := gomail.NewDialer(u.Smtp.Host, u.Smtp.Port, u.Email, u.Password)

	m := gomail.NewMessage()
	m.SetHeader("From", u.Email)
	for _, to := range req.To {
		m.SetAddressHeader("To", to, "")
	}
	builder := new(strings.Builder)
	if _, err := io.Copy(builder, req.Body); err != nil {
		log.Error("cannot copy body to buffer", sl.Err(err), slog.Any("body", req.Body))
		return fmt.Errorf("cannot copy req.Body: %w", err)
	}
	m.SetBody("text/html", builder.String())

	if err := dialer.DialAndSend(m); err != nil {
		log.Error("cannot send message", sl.Err(err), slog.Any("message", m))
		return fmt.Errorf("cannot send", err)
	}

	return nil
}
