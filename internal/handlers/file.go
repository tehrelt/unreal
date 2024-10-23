package handlers

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/authservice"
)

func File(svc *authservice.Service) echo.HandlerFunc {

	return func(c echo.Context) error {

		ctx := c.Request().Context()
		filename := c.Param("filename")

		file, err := svc.File(ctx, filename)
		if err != nil {
			slog.Error("failed to get file", slog.String("id", filename), sl.Err(err))
			return echo.NewHTTPError(echo.ErrBadRequest.Code, "failed to get file")
		}

		body := new(bytes.Buffer)
		if _, err := io.Copy(body, file); err != nil {
			slog.Error("failed to read file", sl.Err(err))
			return err
		}

		ct := http.DetectContentType(body.Bytes())

		return c.Blob(200, ct, body.Bytes())
	}
}
