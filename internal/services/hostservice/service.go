package hostservice

import (
	"context"
	"log/slog"

	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage/models"
)

type HostSaver interface {
	Save(ctx context.Context, in *models.CreateHost) error
}
type FileUploader interface {
	Upload(ctx context.Context, entry *models.File) error
}

type Service struct {
	l        *slog.Logger
	saver    HostSaver
	uploader FileUploader
}

func New(saver HostSaver, uploader FileUploader) *Service {
	return &Service{
		l:        slog.With(sl.Module("hostservice")),
		saver:    saver,
		uploader: uploader,
	}
}
