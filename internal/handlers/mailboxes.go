package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func Mailboxes(ms *mailservice.MailService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, err := extractUser(c)
		if err != nil {
			return err
		}

		mailboxes, err := ms.Mailboxes(ctx)
		if err != nil {
			slog.Error("failed to get mailboxes", sl.Err(err))
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": err.Error(),
			})
		}

		return c.JSON(200, map[string]any{
			"mailboxes": mailboxes,
		})
	}
}
