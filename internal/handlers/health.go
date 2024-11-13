package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func HealthCheck(ms *mailservice.Service) echo.HandlerFunc {

	return func(c echo.Context) error {
		ctx, _, err := extractUser(c)
		if err != nil {
			return err
		}

		info, err := ms.Health(ctx)
		if err != nil {
			slog.Error("health check failed", sl.Err(err))
			return err
		}

		return c.JSON(200, info)
	}
}
