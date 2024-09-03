package app

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/handlers"
	"github.com/tehrelt/unreal/internal/services/authservice"
)

type App struct {
	app    *echo.Echo
	config *config.Config

	as *authservice.AuthService
}

func newApp(cfg *config.Config, as *authservice.AuthService) *App {
	return &App{
		app:    echo.New(),
		config: cfg,
		as:     as,
	}
}

func (a *App) initRoutes() {
	a.app.POST("/login", handlers.LoginHandler(a.as))
}

func (a *App) Run() {

	a.initRoutes()

	port := a.config.Port
	addr := fmt.Sprintf(":%d", port)
	a.app.Logger.Fatal(a.app.Start(addr))
}
