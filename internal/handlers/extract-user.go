package handlers

import (
	"context"
	"log/slog"

	gctx "github.com/tehrelt/unreal/internal/context"
	"github.com/tehrelt/unreal/internal/entity"

	"github.com/labstack/echo/v4"
)

func extractUser(c echo.Context) (context.Context, *entity.SessionInfo, error) {
	session := c.Get("user").(*entity.SessionInfo)
	if session == nil {
		return nil, nil, echo.NewHTTPError(echo.ErrUnauthorized.Code, "no user")
	}

	slog.Debug("extracted user to context", slog.Any("key", gctx.CtxKeyUser))
	return context.WithValue(c.Request().Context(), gctx.CtxKeyUser, session), session, nil
}
