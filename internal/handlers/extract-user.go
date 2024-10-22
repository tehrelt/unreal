package handlers

import (
	"context"
	"log/slog"

	gctx "github.com/tehrelt/unreal/internal/context"

	"github.com/labstack/echo/v4"
)

func extractUser(c echo.Context) (context.Context, error) {
	user := c.Get("user")
	if user == nil {
		return nil, echo.NewHTTPError(echo.ErrUnauthorized.Code, "no user")
	}

	slog.Debug("extracted user to context", slog.Any("key", gctx.CtxKeyUser))
	return context.WithValue(c.Request().Context(), gctx.CtxKeyUser, user), nil
}
