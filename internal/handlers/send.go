package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
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

		slog.Debug("form content", slog.Any("form", form))

		return nil
	}
}
