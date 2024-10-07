package authservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/entity"
)

func (s *AuthService) Authenticate(ctx context.Context, token string) (*entity.SessionInfo, error) {
	claims, err := s.cfg.Jwt.RSA.Verify(token)
	if err != nil {
		return nil, fmt.Errorf("unable to verify token: %w", err)
	}

	slog.Debug("successful verify token", slog.Any("claims", claims))

	info, err := s.sessions.Find(ctx, claims.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to find session: %w", err)
	}

	pass, err := s.decrypt(info.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt session: %w", err)
	}

	info.Password = pass

	slog.Debug("authorized", slog.Any("claims", info))

	return info, nil
}
