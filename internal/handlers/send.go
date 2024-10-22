package handlers

import (
	"log/slog"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func SendMail(ms *mailservice.MailService) echo.HandlerFunc {

	return func(c echo.Context) error {

		ctx, err := extractUser(c)
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
			To:          to,
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
