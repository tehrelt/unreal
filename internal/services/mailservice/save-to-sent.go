package mailservice

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/entity"
	imaps "github.com/tehrelt/unreal/internal/lib/imap"
)

func (ms *Service) saveToSent(ctx context.Context, u *entity.SessionInfo, rawMessage imap.Literal) error {

	log := slog.With(slog.String("Method", "saveToSent"))

	log.Debug("dialing imap", slog.Any("user", u))
	c, cleanup, err := imaps.Dial(u.Email, u.Password, u.Imap.Host, u.Imap.Port)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer cleanup()

	sentFolder, err := ms.findFolderByAttr(ctx, c, "\\Sent")
	if err != nil {
		return fmt.Errorf("failed to find folder: %v", err)
	}

	if _, err := c.Select(sentFolder, false); err != nil {
		return fmt.Errorf("failed to select: %v", err)
	}

	if err := c.Append(sentFolder, []string{imap.SeenFlag}, time.Now(), rawMessage); err != nil {
		return fmt.Errorf("failed to append: %v", err)
	}

	return nil
}
