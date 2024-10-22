package app

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/handlers"
	"github.com/tehrelt/unreal/internal/lib/httpvalidator"
	mw "github.com/tehrelt/unreal/internal/middleware"
	"github.com/tehrelt/unreal/internal/services/authservice"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

type App struct {
	app    *echo.Echo
	config *config.Config

	as *authservice.AuthService
	ms *mailservice.Service
}

func newApp(cfg *config.Config, as *authservice.AuthService, ms *mailservice.Service) *App {
	return &App{
		app:    echo.New(),
		config: cfg,
		as:     as,
		ms:     ms,
	}
}

func (a *App) initRoutes() {

	a.app.Validator = httpvalidator.New(validator.New())

	a.app.Use(emw.Logger())
	a.app.Use(emw.CORSWithConfig(emw.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "http://unreal:3000", "http://10.244.0.13:3000"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE},
		AllowCredentials: true,
	}))

	reqauth := mw.RequireAuth(a.as, a.config)

	a.app.POST("/login", handlers.LoginHandler(a.as))
	a.app.GET("/me", handlers.Profile(a.as), reqauth)
	a.app.GET("/mailboxes", handlers.Mailboxes(a.ms), reqauth)

	mailbox := a.app.Group("/:mailbox", reqauth)
	mailbox.GET("", handlers.Messages(a.ms))
	mailbox.GET("/:mailnum", handlers.Message(a.ms))

	a.app.GET("/attachment/:filename", handlers.Attachment(a.ms), reqauth)
	a.app.POST("/send", handlers.SendMail(a.ms), reqauth)
}

func (a *App) Run() {

	a.initRoutes()

	host := a.config.Host

	a.app.Logger.Fatal(a.app.Start(host))
}
