package authservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (s *AuthService) Profile(ctx context.Context, email string) (*entity.User, error) {
	fn := "authservice.Profile"
	log := s.logger.With(sl.Method(fn))

	log.Debug("find user by email", slog.String("email", email))
	user, err := s.userProvider.Find(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	out := &entity.User{
		Email:   user.Email,
		Name:    user.Name,
		Picture: user.ProfilePicture,
	}

	return out, nil
}
