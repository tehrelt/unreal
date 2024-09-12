package handlers

import (
	"log/slog"
	"net/http"

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

		c.SetCookie(&http.Cookie{
			Name:     "token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteDefaultMode,
		})

		return c.JSON(200, &response{
			Token: token,
		})
	}
}
