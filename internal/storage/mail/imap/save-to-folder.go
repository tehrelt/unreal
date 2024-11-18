package imap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (r *Repository) saveMessageToFolder(ctx context.Context, folder string, msg imap.Literal) error {

	fn := "imap.saveMessageToFolder"
	log := r.l.With(sl.Method(fn))

	c, err := r.ctxman.get(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if _, err := c.Select(folder, false); err != nil {
		return fmt.Errorf("failed to select: %v", err)
	}

	log.Debug("appending message to folder", slog.String("folder", folder))
	if err := c.Append(folder, []string{imap.SeenFlag}, time.Now(), msg); err != nil {
		return fmt.Errorf("failed to append: %v", err)
	}

	return nil
}

func (r *Repository) SaveSentMessage(ctx context.Context, msg io.Reader) error {

	fn := "imap.SaveSentMessage"

	folder, err := r.findFolderByAttr(ctx, "\\Sent")
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return r.saveMessageToFolder(ctx, folder, msg.(imap.Literal))
}

func (r *Repository) SaveDraftMessage(ctx context.Context, msg io.Reader) error {

	fn := "imap.SaveDraft"

	folder, err := r.findFolderByAttr(ctx, "\\Drafts")
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return r.saveMessageToFolder(ctx, folder, msg.(imap.Literal))
}
