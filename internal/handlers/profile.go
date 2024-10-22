package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/services/authservice"
)

type ProfileResponse struct {
	Email   string  `json:"email"`
	Name    *string `json:"name"`
	Picture *string `json:"picture"`
}

func Profile(as *authservice.AuthService) echo.HandlerFunc {
	type response struct {
		Email string `json:"email"`
	}
	return func(c echo.Context) error {

		ctx, err := extractUser(c)
		if err != nil {
			return err
		}
		u := ctx.Value(context.CtxKeyUser).(*entity.SessionInfo)

		user, err := as.Profile(ctx, u.Email)
		if err != nil {
			return echo.NewHTTPError(500, err.Error())
		}

		return c.JSON(200, &ProfileResponse{
			Email:   user.Email,
			Name:    user.Name,
			Picture: user.Picture,
		})
	}
}
