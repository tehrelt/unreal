package authservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/aes"
	"github.com/tehrelt/unreal/internal/lib/imap"
)

func (s *AuthService) Login(ctx context.Context, d *dto.LoginDto) (string, error) {

	_, cleanup, err := imap.Dial(d.Email, d.Password, d.Imap.Host, d.Imap.Port)
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
		Password: pass,
		Imap:     d.Imap,
		Smtp:     d.Smtp,
	}

	slog.Debug("signing token with RSA")
	token, err := s.cfg.Jwt.RSA.Sign(claims, s.cfg.Jwt.Ttl)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	slog.Info("signed token", slog.String("token", token))

	return token, nil
}
