package authservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/aes"
)

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
