package authservice

import (
	"context"

	"github.com/tehrelt/unreal/internal/storage/models"
)

func (s *AuthService) File(ctx context.Context, id string) (*models.File, error) {
	return s.fileProvider.File(ctx, id)
}
