package handlers

import (
	"context"

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

func Messages(ms *mailservice.MailService) echo.HandlerFunc {

	return func(c echo.Context) error {

		var request MessagesRequest

		user := c.Get("user")
		if user == nil {
			return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
				"error": "no user in context",
			})
		}

		mailbox := c.Param("mailbox")
		if mailbox == "" {
			return c.JSON(echo.ErrBadRequest.Code, map[string]any{
				"error": "no mailbox in path",
			})
		}

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

		out, err := ms.Messages(
			context.WithValue(c.Request().Context(), "user", user),
			&dto.FetchMessagesDto{
				Mailbox: entity.NewMailboxName(mailbox),
				Limit: func() int {
					if request.Limit == 0 {
						return defaultLimit
					}
					return request.Limit
				}(),
				Page: request.Page,
			},
		)
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
