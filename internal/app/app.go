package app

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/config"
)

type App struct {
	app    *echo.Echo
	config *config.Config
}

func newApp(cfg *config.Config) *App {
	return &App{
		app:    echo.New(),
		config: cfg,
	}
}

func (a *App) Run() {
	port := a.config.Port
	addr := fmt.Sprintf(":%d", port)
	a.app.Logger.Fatal(a.app.Start(addr))
}
