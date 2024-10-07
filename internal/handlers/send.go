package handlers

import (
	"context"
	"log/slog"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func SendMail(ms *mailservice.MailService) echo.HandlerFunc {

	return func(c echo.Context) error {
		user := c.Get("user")
		if user == nil {
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": "no user in context",
			})
		}

		form, err := c.MultipartForm()
		if err != nil {
			return c.JSON(500, map[string]any{
				"error": err.Error(),
			})
		}

		ctx := c.Request().Context()

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

		if err := ms.Send(context.WithValue(ctx, "user", user), req); err != nil {
			return c.JSON(500, map[string]any{
				"error": err.Error(),
			})
		}

		return nil
	}
}
