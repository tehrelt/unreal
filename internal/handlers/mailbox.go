package handlers

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func Mailbox(ms *mailservice.MailService) echo.HandlerFunc {

	type response struct {
		Messages []*entity.Message `json:"messages"`
		Total    int               `json:"total"`
	}

	return func(c echo.Context) error {

		user := c.Get("user")
		if user == nil {
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": "no user in context",
			})
		}

		mailbox := c.Param("mailbox")

		messages, total, err := ms.Messages(
			context.WithValue(c.Request().Context(), "user", user),
			entity.NewMailboxName(mailbox),
		)
		if err != nil {
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": err.Error(),
			})
		}

		return c.JSON(200, &response{
			Messages: messages,
			Total:    total,
		})
	}
}
