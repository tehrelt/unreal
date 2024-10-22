package handlers

import (
	"log/slog"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func SendMail(ms *mailservice.Service) echo.HandlerFunc {

	return func(c echo.Context) error {

		ctx, u, err := extractUser(c)
		if err != nil {
			return err
		}

		form, err := c.MultipartForm()
		if err != nil {
			return c.JSON(500, map[string]any{
				"error": err.Error(),
			})
		}

		to := form.Value["to"]
		subject := form.Value["subject"][0]
		body := strings.NewReader(form.Value["body"][0])
		attachments := form.File["attachment"]
		embedded := form.File["embedded"]

		slog.Debug(
			"form content",
			slog.Any("to", to),
			slog.String("subject", subject),
			slog.Any("body", body.Len()),
			slog.Any("attachments", attachments),
			slog.Any("embedded", embedded),
		)

		req := &dto.SendMessageDto{
			From: &dto.MailRecord{Email: u.Email},
			To: lo.Map(to, func(email string, _ int) *dto.MailRecord {
				return &dto.MailRecord{Email: email}
			}),
			Subject:     subject,
			Body:        body,
			Attachments: attachments,
		}

		if err := ms.Send(ctx, req); err != nil {
			return echo.NewHTTPError(500, err.Error())
		}

		return nil
	}
}
