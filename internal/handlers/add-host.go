package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/services/hostservice"
)

func AddHost(s *hostservice.Service) echo.HandlerFunc {
	return func(c echo.Context) error {

		form, err := c.MultipartForm()
		if err != nil {
			return err
		}

		hosts := form.Value["host"]
		if len(hosts) == 0 {
			return echo.NewHTTPError(400, "host is required")
		}

		files := form.File["picture"]
		if len(files) == 0 {
			return echo.NewHTTPError(400, "picture is required")
		}

		host := hosts[0]
		file := files[0]

		ctx := c.Request().Context()

		if err := s.Add(ctx, host, file); err != nil {
			return err
		}

		return nil
	}
}
