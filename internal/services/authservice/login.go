package authservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/imap"
)

func (s *AuthService) Login(ctx context.Context, in *dto.LoginDto) (string, error) {

	_, cleanup, err := imap.Dial(in.Email, in.Password, in.Imap.Host, in.Imap.Port)
	if err != nil {
		return "", err
	}
	defer cleanup()

	pass, err := s.encrypt(in.Password)
	if err != nil {
		return "", fmt.Errorf("unable to encrypt password: %w", err)
	}

	info := &entity.SessionInfo{
		Email:    in.Email,
		Password: pass,
		Imap:     in.Imap,
		Smtp:     in.Smtp,
	}

	slog.Debug("creating session", slog.Any("session", info))

	id, err := s.sessions.Save(ctx, info, s.cfg.Jwt.Ttl)
	if err != nil {
		return "", fmt.Errorf("unable to save session: %w", err)
	}

	slog.Debug("signing token with RSA")
	token, err := s.cfg.Jwt.RSA.Sign(&entity.Claims{Id: id}, s.cfg.Jwt.Ttl)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	slog.Info("signed token", slog.String("token", token))

	return token, nil
}
