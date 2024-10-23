package authservice

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage/models"
)

func (s *AuthService) UpdateUser(ctx context.Context, in *entity.UpdateUser) error {
	fn := "authservice.UpdateUser"
	log := s.logger.With(sl.Method(fn))

	var filename *string
	if in.Picture != nil {

		pic, err := in.Picture.Open()
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}
		defer pic.Close()

		id := uuid.New().String()
		ext := filepath.Ext(in.Picture.Filename)
		name := id + ext

		if err := s.fileUploader.Upload(ctx, &models.File{
			Filename:    name,
			ContentType: in.Picture.Header.Get("Content-Type"),
			Reader:      pic,
		}); err != nil {
			log.Error("failed to upload picture")
			return fmt.Errorf("%s: %w", fn, err)
		}

		filename = &name
	}

	model := &models.UpdateUser{
		UserBase: models.UserBase{
			Email: in.Email,
		},
		Name:           in.Name,
		ProfilePicture: filename,
	}

	if err := s.userUpdater.Update(ctx, model); err != nil {
		log.Error("failed to update user", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
