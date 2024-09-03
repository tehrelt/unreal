package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/authservice"
)

func LoginHandler(as *authservice.AuthService) echo.HandlerFunc {

	type response struct {
		Token string `json:"token"`
	}

	return func(c echo.Context) error {

		req := new(dto.LoginDto)

		if err := c.Bind(req); err != nil {
			slog.Error("failed to bind:", sl.Err(err))
			return echo.ErrInternalServerError
		}

		token, err := as.Login(c.Request().Context(), req)
		if err != nil {
			slog.Error("failed to login:", sl.Err(err))
			return echo.ErrInternalServerError
		}

		return c.JSON(200, &response{
			Token: token,
		})
	}
}
