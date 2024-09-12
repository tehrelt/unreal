package handlers

import (
	"context"
	"io"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func Attachment(ms *mailservice.MailService) echo.HandlerFunc {

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

		mailbox := c.QueryParam("mailbox")
		if mailbox == "" {
			return c.JSON(echo.ErrBadRequest.Code, map[string]any{
				"error": "emptry mailbox",
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

		filename := c.Param("filename")
		if filename == "" {
			return c.JSON(echo.ErrBadRequest.Code, map[string]any{
				"error": "emptry cid",
			})
		}

		reader, ct, err := ms.GetAttachment(context.WithValue(c.Request().Context(), "user", user), mailbox, uint32(inum), filename)
		if err != nil {
			slog.Error("failed to get mail", sl.Err(err))
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": err.Error(),
			})
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			slog.Error("failed to read attachment", sl.Err(err))
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": err.Error(),
			})
		}

		slog.Info("got attachment", slog.String("filename", filename), slog.Int("size", len(body)), slog.String("ct", ct))

		return c.Blob(200, ct, body)
	}
}
