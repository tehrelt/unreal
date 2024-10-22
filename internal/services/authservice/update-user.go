package authservice

import (
	"context"
	"fmt"

	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage/models"
)

func (s *AuthService) UpdateUser(ctx context.Context, in *entity.UpdateUser) error {
	fn := "authservice.UpdateUser"
	log := s.logger.With(sl.Method(fn))

	model := &models.UpdateUser{
		UserBase: models.UserBase{
			Email: in.Email,
		},
		Name: in.Name,
	}

	if err := s.userUpdater.Update(ctx, model); err != nil {
		log.Error("failed to update user", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
