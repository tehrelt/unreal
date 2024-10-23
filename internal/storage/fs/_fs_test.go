// go:build
package fs_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/fs"
	"gotest.tools/v3/assert"
)

const staticpath = "./.volume"

func id() string {
	provider := uuid.New()
	return provider.Provide()
}

func suite(t *testing.T) (context.Context, *fs.FileStorage) {
	t.Helper()

	cfg := conf()
	fs := fs.New(cfg)
	ctx := context.Background()
	return ctx, fs
}

func conf() *config.Config {
	cfg := &config.Config{
		FS: config.FileStorage{
			Path: staticpath,
		},
	}
	return cfg
}

func TestUploadRead(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		t.Run("UploadThenRead", func(t *testing.T) {
			ctx, fs := suite(t)

			filename := id()
			content := fmt.Sprintf("lorem ipsum - %d", i)

			err := fs.Upload(ctx, domain.NewEntry(filename, strings.NewReader(content)))
			require.NoError(t, err)
			t.Cleanup(func() {
				os.Remove(fmt.Sprintf("%s/%s", staticpath, filename))
			})

			e, err := fs.File(ctx, filename)
			require.NoError(t, err)

			read, err := io.ReadAll(e)
			require.NoError(t, err)

			assert.Equal(t, filename, e.Filename())
			assert.Equal(t, content, string(read))
		})
	}
}

func TestReadNotExistsFile(t *testing.T) {
	t.Parallel()

	ctx, fs := suite(t)
	filename := id()

	_, err := fs.File(ctx, filename)
	require.ErrorIs(t, err, storage.ErrFileNotExists)
}

func TestUploadExistingFile(t *testing.T) {
	t.Parallel()

	ctx, fs := suite(t)

	filename := "exist.txt"
	content := "exist1"

	if err := fs.Upload(ctx, domain.NewEntry(filename, strings.NewReader(content))); err != nil {
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		os.Remove(fmt.Sprintf("%s/%s", staticpath, filename))
	})

	content = "exist2"
	if err := fs.Upload(ctx, domain.NewEntry(filename, strings.NewReader(content))); err != nil {
		require.ErrorIs(t, err, storage.ErrFileAlreadyExists)
	}
}

func TestUploadMany(t *testing.T) {
	t.Parallel()

	ctx, fs := suite(t)

	create := func(filename, content string) *domain.Entry {
		return domain.NewEntry(filename, strings.NewReader(content))
	}

	contents := make([]string, 0, 10)
	files := make([]*domain.Entry, 0, 10)
	for range 10 {
		filename, content := id(), id()

		files = append(files, create(filename, content))
		contents = append(contents, content)

		t.Cleanup(func() {
			os.Remove(fmt.Sprintf("%s/%s", staticpath, filename))
		})
	}

	if err := fs.UploadMany(ctx, files); err != nil {
		require.NoError(t, err)
	}

	for i := range 10 {
		file, err := fs.File(ctx, files[i].Filename())
		require.NoError(t, err)

		got, err := io.ReadAll(file)
		require.NoError(t, err)

		want := contents[i]

		assert.Equal(t, want, string(got))
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	ctx, fs := suite(t)

	filename := id()
	content := id()
	if err := fs.Upload(ctx, domain.NewEntry(filename, strings.NewReader(content))); err != nil {
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		os.Remove(fmt.Sprintf("%s/%s", staticpath, filename))
	})

	if err := fs.Delete(ctx, filename); err != nil {
		require.NoError(t, err)
	}
}

func TestDeleteNotExistFile(t *testing.T) {
	t.Parallel()

	ctx, fs := suite(t)
	filename := id()

	if err := fs.Delete(ctx, filename); err != nil {
		require.NoError(t, err)
	}
}
