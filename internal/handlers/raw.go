package handlers

import (
	"io"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func Raw(ms *mailservice.Service) echo.HandlerFunc {

	return func(c echo.Context) error {

		ctx, _, err := extractUser(c)
		if err != nil {
			return err
		}

		var request MessageRequest

		if err := echo.PathParamsBinder(c).
			String("mailbox", &request.Mailbox).
			Int("mailnum", &request.Mailnum).
			BindError(); err != nil {
			return echo.NewHTTPError(echo.ErrBadRequest.Code, err.Error())
		}

		if err := c.Validate(request); err != nil {
			return err
		}

		raw, err := ms.Raw(
			ctx,
			request.Mailbox,
			uint32(request.Mailnum),
		)
		if err != nil {
			slog.Error("failed to get mail", sl.Err(err))
			return echo.NewHTTPError(echo.ErrInternalServerError.Code, err.Error())
		}

		content, err := io.ReadAll(raw)
		if err != nil {
			slog.Error("failed to read mail", sl.Err(err))
			return echo.NewHTTPError(echo.ErrInternalServerError.Code, err.Error())
		}

		return c.Blob(200, "text/plain", content)
	}

}
