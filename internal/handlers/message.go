package handlers

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func Message(ms *mailservice.MailService) echo.HandlerFunc {

	type response struct {
		Mail *entity.MessageWithBody `json:"mail"`
	}

	return func(c echo.Context) error {

		user := c.Get(string("user"))
		if user == nil {
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": "no user in context",
			})
		}

		mailbox := c.Param("mailbox")
		if mailbox == "" {
			return c.JSON(echo.ErrBadRequest.Code, map[string]any{
				"error": "empty mailbox",
			})
		}

		num := c.QueryParam("mailnum")
		if num == "" {
			return c.JSON(echo.ErrBadRequest.Code, map[string]any{
				"error": "emptry mail number",
			})
		}

		inum, err := strconv.Atoi(num)
		if err != nil {
			return c.JSON(echo.ErrBadRequest.Code, map[string]any{
				"error": "invalid mail number",
			})
		}

		if inum < 0 {
			return c.JSON(echo.ErrBadRequest.Code, map[string]any{
				"error": "mail number must be positive",
			})
		}

		mail, err := ms.Message(context.WithValue(c.Request().Context(), "user", user), entity.MailboxName(mailbox), uint32(inum))
		if err != nil {
			slog.Error("failed to get mail", sl.Err(err))
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": err.Error(),
			})
		}

		return c.JSON(200, &response{Mail: mail})
	}

}
