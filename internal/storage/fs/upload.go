package fs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/models"
)

// Upload implements fileservice.FileUploader.
func (f *FileStorage) Upload(ctx context.Context, entry *models.File) error {
	fn := "fs.Upload"
	log := f.logger.With(sl.Method(fn), slog.String("filename", entry.Filename))

	log.Debug("checking volume")
	if err := f.checkVolume(); err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return f.upload(ctx, entry)
}

func (f *FileStorage) upload(_ context.Context, entry *models.File) error {
	fn := "fs.Upload"
	filename := entry.Filename
	log := f.logger.With(sl.Method(fn), slog.String("filename", filename))

	path := f.joinFileName(filename)

	log.Debug("opening file", slog.String("path", path))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%s: %w", fn, storage.ErrFileAlreadyExists)
		}
	}
	defer file.Close()

	log.Debug("start write file")
	start := time.Now()
	if _, err := io.Copy(file, entry); err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	log.Debug("file written", sl.Millis("took", time.Since(start)))

	return nil
}
