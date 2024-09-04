package handlers

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func Mailboxes(ms *mailservice.MailService) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user")
		if user == nil {
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": "no user in context",
			})
		}

		mailboxes, err := ms.Mailboxes(context.WithValue(c.Request().Context(), "user", user))
		if err != nil {
			slog.Error("failed to get mailboxes", sl.Err(err))
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": err.Error(),
			})
		}

		slog.Debug("got mailboxes", slog.Any("mailboxes", mailboxes))

		return c.JSON(200, map[string]any{
			"mailboxes": mailboxes,
		})
	}
}
