package fs

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

type FileStorage struct {
	staticPath string
	logger     *slog.Logger
}

func (f *FileStorage) createVolume() error {
	f.logger.Debug("creating volume", slog.String("staticPath", f.staticPath))
	return os.Mkdir(f.staticPath, 0755)
}

func (f *FileStorage) checkVolume() error {
	fn := "fs.checkVolume"
	log := f.logger.With(sl.Method(fn))

	stat, err := os.Stat(f.staticPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn("volume does not exists", slog.String("staticPath", f.staticPath))
			return f.createVolume()
		}

		log.Error("failed to stat path", slog.String("staticPath", f.staticPath), sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	if !stat.IsDir() {
		log.Error("current path to static points to file not a directory", slog.String("staticPath", f.staticPath))
		panic(fmt.Errorf("current path to static points to file not a directory"))
	}

	return nil
}

func (f *FileStorage) joinFileName(filename string) string {
	return fmt.Sprintf("%s/%s", f.staticPath, filename)
}

func New(cfg *config.Config) *FileStorage {
	fs := &FileStorage{
		staticPath: cfg.Fs.StaticPath,
		logger:     slog.With(sl.Module("fs.FileStorage")),
	}

	_ = fs.checkVolume()

	return fs
}
