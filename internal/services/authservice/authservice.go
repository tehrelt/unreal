package authservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/aes"
	"github.com/tehrelt/unreal/internal/lib/imap"
)

type AuthService struct {
	cfg *config.Config
}

func New(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (*entity.Claims, error) {

	claims, err := s.cfg.Jwt.RSA.Verify(token)
	if err != nil {
		return nil, fmt.Errorf("unable to verify token: %w", err)
	}

	slog.Debug("successful verify token", slog.Any("claims", claims))

	pass, err := aes.Decrypt(s.cfg.AES.Secret, claims.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt password: %w", err)
	}

	claims.Password = string(pass)

	slog.Debug("authorized", slog.Any("claims", claims))

	return claims, nil
}

func (s *AuthService) Login(ctx context.Context, d *dto.LoginDto) (string, error) {

	_, cleanup, err := imap.Dial(d.Email, d.Password, d.Host, d.Port)
	if err != nil {
		return "", err
	}
	defer cleanup()

	slog.Debug("dial successfull")

	pass, err := aes.Encrypt(s.cfg.AES.Secret, d.Password)
	if err != nil {
		return "", fmt.Errorf("unable to encrypt password: %w", err)
	}

	claims := &entity.Claims{
		Email:    d.Email,
		Host:     d.Host,
		Password: pass,
		Port:     d.Port,
	}

	slog.Debug("signing token with RSA")
	token, err := s.cfg.Jwt.RSA.Sign(claims, s.cfg.Jwt.Ttl)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	slog.Info("signed token", slog.String("token", token))

	return token, nil
}
