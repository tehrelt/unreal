package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

type MessageRequest struct {
	Mailbox string `param:"mailbox" validate:"required"`
	Mailnum int    `param:"mailnum" validate:"gte=0"`
}
type MessageResponse struct {
	Mail *entity.MessageWithBody `json:"mail"`
}

func Message(ms *mailservice.MailService) echo.HandlerFunc {

	return func(c echo.Context) error {

		ctx, err := extractUser(c)
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

		mail, err := ms.Message(
			ctx,
			request.Mailbox,
			uint32(request.Mailnum),
		)
		if err != nil {
			slog.Error("failed to get mail", sl.Err(err))
			return echo.NewHTTPError(echo.ErrInternalServerError.Code, err.Error())
		}

		return c.JSON(200, &MessageResponse{Mail: mail})
	}

}
