package app

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
	"github.com/samber/lo"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/handlers"
	"github.com/tehrelt/unreal/internal/lib/httpvalidator"
	mw "github.com/tehrelt/unreal/internal/middleware"
	"github.com/tehrelt/unreal/internal/services/authservice"
	"github.com/tehrelt/unreal/internal/services/hostservice"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

type App struct {
	app    *echo.Echo
	config *config.Config

	as *authservice.Service
	ms *mailservice.Service
	hs *hostservice.Service
}

func newApp(cfg *config.Config, as *authservice.Service, ms *mailservice.Service, hs *hostservice.Service) *App {
	return &App{
		app:    echo.New(),
		config: cfg,
		as:     as,
		ms:     ms,
		hs:     hs,
	}
}

func (a *App) initRoutes() {

	a.app.Validator = httpvalidator.New(validator.New())

	origins := lo.Map(strings.Split(a.config.CORS.AllowOrigins, ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})

	slog.Info("CORS origins", slog.Any("origins", origins))

	a.app.Use(emw.Logger())
	a.app.Use(emw.CORSWithConfig(emw.CORSConfig{
		AllowCredentials: true,
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE},
		AllowOrigins:     origins,
	}))

	reqauth := mw.RequireAuth(a.as, a.config)

	a.app.POST("/login", handlers.LoginHandler(a.as))
	a.app.GET("/me", handlers.Profile(a.as), reqauth)
	a.app.PUT("/me", handlers.UpdateProfile(a.as), reqauth)
	a.app.GET("/health", handlers.HealthCheck(a.ms), reqauth)
	a.app.GET("/mailboxes", handlers.Mailboxes(a.ms), reqauth)
	a.app.GET("/file/:filename", handlers.File(a.as))

	mailbox := a.app.Group("/:mailbox", reqauth)
	mailbox.GET("", handlers.Messages(a.ms))
	mailbox.GET("/:mailnum", handlers.Message(a.ms))
	mailbox.GET("/:mailnum/raw", handlers.Raw(a.ms))
	mailbox.DELETE("/:mailnum", handlers.Delete(a.ms))

	a.app.GET("/attachment/:filename", handlers.Attachment(a.ms), reqauth)
	a.app.POST("/send", handlers.SendMail(a.ms), reqauth)
	a.app.POST("/draft", handlers.Draft(a.ms), reqauth)

	hosts := a.app.Group("/hosts")
	hosts.POST("/", handlers.AddHost(a.hs))
}

func (a *App) Run() {

	a.initRoutes()

	host := a.config.Hostname
	port := a.config.Port
	addr := fmt.Sprintf("%s:%d", host, port)

	a.app.Logger.Fatal(a.app.Start(addr))
}
