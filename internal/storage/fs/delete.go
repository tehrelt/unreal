package fs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (f *FileStorage) Delete(ctx context.Context, filename string) error {
	fn := "fs.Delete"
	log := f.logger.With(sl.Method(fn), slog.String("filename", filename))

	path := f.joinFileName(filename)

	log.Debug("start delete file")
	start := time.Now()
	if err := os.Remove(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Error("cannot delete file", sl.Err(err))
			return fmt.Errorf("%s: %w", fn, err)
		}
	}
	log.Debug("file deleted", sl.Millis("took", time.Since(start)))

	return nil
}
