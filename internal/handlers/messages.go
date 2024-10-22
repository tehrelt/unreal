package handlers

import (
	"log/slog"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

const defaultLimit = 50

type MessagesRequest struct {
	Limit int `query:"limit,omitempty" validate:"omitempty,gt=0"`
	Page  int `query:"page,omitempty" validate:"omitempty,gt=0"`
}

type MessagesResponse struct {
	Messages []entity.Message `json:"messages"`
	HasNext  *int             `json:"hasNext"`
	Total    int              `json:"total"`
}

func Messages(ms *mailservice.Service) echo.HandlerFunc {

	return func(c echo.Context) error {

		var request MessagesRequest

		ctx, err := extractUser(c)
		if err != nil {
			return err
		}

		mailboxescaped := c.Param("mailbox")
		if mailboxescaped == "" {
			return c.JSON(echo.ErrBadRequest.Code, map[string]any{
				"error": "no mailbox in path",
			})
		}

		mailbox, err := url.QueryUnescape(mailboxescaped)
		if err != nil {
			return echo.NewHTTPError(500, err.Error())
		}

		slog.Debug(
			"/messages request",
			slog.String("mailbox", mailbox),
			slog.Int("limit", request.Limit),
			slog.Int("page", request.Page),
		)

		if err := echo.QueryParamsBinder(c).
			Int("limit", &request.Limit).
			Int("page", &request.Page).
			BindError(); err != nil {
			return c.JSON(echo.ErrBadRequest.Code, map[string]any{
				"error": err.Error(),
			})
		}

		if err := c.Validate(request); err != nil {
			return c.JSON(echo.ErrBadRequest.Code, map[string]any{
				"error": err.Error(),
			})
		}

		in := &dto.FetchMessagesDto{
			Mailbox: mailbox,
			Limit: func() int {
				if request.Limit == 0 {
					return defaultLimit
				}
				return request.Limit
			}(),
			Page: request.Page,
		}

		out, err := ms.Messages(ctx, in)
		if err != nil {
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": err.Error(),
			})
		}

		return c.JSON(200, &MessagesResponse{
			Messages: out.Messages,
			HasNext: func() *int {
				if out.HasNextPage {
					p := request.Page + 1
					return &p
				}
				return nil
			}(),
			Total: out.Total,
		})
	}
}
