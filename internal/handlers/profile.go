package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/entity"
)

func Profile() echo.HandlerFunc {
	type response struct {
		Email string `json:"email"`
		// Host  string `json:"host"`
	}
	return func(c echo.Context) error {
		u := c.Get("user").(*entity.Claims)
		if u == nil {
			slog.Warn("no user in context")
			return c.JSON(401, map[string]any{
				"error": "unathorized",
			})
		}

		return c.JSON(200, &response{
			Email: u.Email,
		})
	}
}
