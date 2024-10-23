package fs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/models"
)

// File implements fileservice.FileProvider.
func (f *FileStorage) File(ctx context.Context, filename string) (*models.File, error) {
	fn := "fs.File"
	log := f.logger.With(sl.Method(fn), slog.String("filename", filename))

	path := f.joinFileName(filename)

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%s: %w", fn, storage.ErrFileNotExists)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	log.Debug("start file read")
	start := time.Now()
	if _, err := io.Copy(buf, file); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	log.Debug("file read", sl.Millis("took", time.Since(start)))

	return models.NewFile(buf, filename, http.DetectContentType(buf.Bytes())), nil
}
