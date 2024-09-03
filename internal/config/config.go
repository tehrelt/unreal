package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/tehrelt/unreal/internal/lib/jwt"
	"github.com/tehrelt/unreal/internal/lib/logger/prettyslog"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

type Config struct {
	Env  string `env:"ENV" env-default:"local"`
	Port int    `env:"PORT" env-required:"true" env-default:"4200"`

	App struct {
		Name    string `env:"APP_NAME" env-required:"true" env-default:"unreal"`
		Version string `env:"APP_VERSION" env-required:"true" env-default:"0.0.1"`
	}

	Cert struct {
		PrivateKeyFile string `env:"CERT_PRIVATE_KEY_FILE" env-required:"true" env-default:"./cert/id_rsa"`
		PublicKeyFile  string `env:"CERT_PUBLIC_KEY_FILE" env-required:"true" env-default:"./cert/id_rsa.pub"`
	}

	AES struct {
		Secret string `env:"AES_SECRET"`
	}

	Jwt struct {
		RSA       *jwt.JWT
		Ttl       time.Duration
		TtlString string `env:"JWT_TTL" env-required:"true" env-default:"10m"`
	}
}

func New() *Config {
	config := new(Config)

	if err := cleanenv.ReadEnv(config); err != nil {
		slog.Error("error when reading env", sl.Err(err))
		header := fmt.Sprintf("%s - %s", os.Getenv("APP_NAME"), os.Getenv("APP_VERSION"))
		usage := cleanenv.FUsage(os.Stdout, config, &header)
		usage()
		os.Exit(-1)
	}

	config.setupLogger()
	if err := config.setupJwt(); err != nil {
		slog.Error("error when reading jwt certificates", sl.Err(err))
		panic(err)
	}
	if err := config.parseTtl(); err != nil {
		slog.Error("error when parsing jwt ttl", sl.Err(err))
		panic(err)
	}

	slog.Info("config setup", slog.Any("c", config))

	return config
}

func (c *Config) parseTtl() error {
	var err error
	c.Jwt.Ttl, err = time.ParseDuration(c.Jwt.TtlString)
	if err != nil {
		return fmt.Errorf("unable to parse duration %s: %w", c.Jwt.TtlString, err)
	}
	return nil
}

func (c *Config) setupJwt() error {

	private, err := os.ReadFile(c.Cert.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("unable to read file %s: %w", c.Cert.PrivateKeyFile, err)
	}

	public, err := os.ReadFile(c.Cert.PublicKeyFile)
	if err != nil {
		return fmt.Errorf("unable to read file %s: %w", c.Cert.PublicKeyFile, err)
	}

	c.Jwt.RSA = jwt.NewJWT(private, public)

	return nil
}

func (cfg *Config) setupLogger() {
	var log *slog.Logger
	switch cfg.Env {
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "dev":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		log = slog.New(prettyslog.NewPrettyHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(log)
}
