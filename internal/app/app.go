package app

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/handlers"
	"github.com/tehrelt/unreal/internal/middleware"
	"github.com/tehrelt/unreal/internal/services/authservice"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

type App struct {
	app    *echo.Echo
	config *config.Config

	as *authservice.AuthService
	ms *mailservice.MailService
}

func newApp(cfg *config.Config, as *authservice.AuthService, ms *mailservice.MailService) *App {
	return &App{
		app:    echo.New(),
		config: cfg,
		as:     as,
		ms:     ms,
	}
}

func (a *App) initRoutes() {
	a.app.POST("/login", handlers.LoginHandler(a.as))
	a.app.GET("/mailboxes", handlers.Mailboxes(a.ms), middleware.RequireAuth(a.as, a.config))
}

func (a *App) Run() {

	a.initRoutes()

	port := a.config.Port
	addr := fmt.Sprintf(":%d", port)
	a.app.Logger.Fatal(a.app.Start(addr))
}
