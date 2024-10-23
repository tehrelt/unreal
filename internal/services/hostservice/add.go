package hostservice

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage/models"
)

func (s *Service) Add(ctx context.Context, host string, picture *multipart.FileHeader) error {

	fn := "hostservice.Add"
	log := s.l.With(sl.Method(fn))

	log.Debug("uploading picture")
	pic, err := picture.Open()
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer pic.Close()

	id := uuid.New().String()
	ext := filepath.Ext(picture.Filename)
	filename := id + ext
	file := models.NewFile(pic, filename, picture.Header.Get("Content-Type"))

	if err := s.uploader.Upload(ctx, file); err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	log.Debug("adding host")
	if err := s.saver.Save(ctx, &models.CreateHost{Host: host, Picture: filename}); err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
