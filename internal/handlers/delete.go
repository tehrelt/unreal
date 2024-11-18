package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func Delete(ms *mailservice.Service) echo.HandlerFunc {

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

		if err := ms.Delete(
			ctx,
			request.Mailbox,
			uint32(request.Mailnum),
		); err != nil {
			slog.Error("failed to delete message", sl.Err(err))
			return echo.NewHTTPError(echo.ErrInternalServerError.Code, err.Error())
		}

		return c.NoContent(200)
	}

}
