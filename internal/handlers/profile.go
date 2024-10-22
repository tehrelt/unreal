package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/entity"
)

func Profile() echo.HandlerFunc {
	type response struct {
		Email string `json:"email"`
	}
	return func(c echo.Context) error {

		ctx, err := extractUser(c)
		if err != nil {
			return err
		}

		u := ctx.Value(context.CtxKeyUser).(*entity.SessionInfo)
		return c.JSON(200, &response{
			Email: u.Email,
		})
	}
}
