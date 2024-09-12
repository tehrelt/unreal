package middleware

import (
	"log/slog"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services/authservice"
)

func RequireAuth(as *authservice.AuthService, cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			slog.Debug("require auth check")

			var token string

			ctx := c.Request().Context()

			cookieToken, err := c.Cookie("token")
			if cookieToken != nil {
				token = cookieToken.Value
				slog.Debug("get token from cookie", slog.String("token", token))
			} else {
				authHeader := c.Request().Header.Get("Authorization")
				if authHeader == "" {
					return c.String(401, "Unauthorized")
				}

				tok := strings.Split(authHeader, " ")
				if len(tok) < 2 {
					return c.String(401, "Unauthorized")
				}

				token = tok[1]
			}

			slog.Debug("get token", slog.String("token", token))

			user, err := as.Authenticate(ctx, token)
			if err != nil {
				slog.Error("failed to authenticate token", sl.Err(err))
				return c.JSON(echo.ErrInternalServerError.Code, map[string]any{
					"error": err.Error(),
				})
			}

			c.Set("user", user)

			return next(c)
		}
	}
}
