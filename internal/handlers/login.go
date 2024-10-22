package handlers

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/authservice"
)

type LoginResponse struct {
	Token      string `json:"token"`
	FirstLogon bool   `json:"firstLogon"`
}

func LoginHandler(as *authservice.AuthService) echo.HandlerFunc {

	return func(c echo.Context) error {

		req := new(dto.LoginDto)

		if err := c.Bind(req); err != nil {
			slog.Error("failed to bind:", sl.Err(err))
			return echo.ErrInternalServerError
		}

		res, err := as.Login(c.Request().Context(), req)
		if err != nil {
			slog.Error("failed to login:", sl.Err(err))
			return echo.ErrInternalServerError
		}

		c.SetCookie(&http.Cookie{
			Name:     "token",
			Value:    res.Token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteDefaultMode,
		})

		return c.JSON(200, &LoginResponse{
			Token:      res.Token,
			FirstLogon: res.FirstLogon,
		})
	}
}
