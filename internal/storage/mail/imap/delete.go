package imap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (r *Repository) Delete(ctx context.Context, mailbox string, num uint32) error {

	fn := "mail.Delete"
	log := slog.With(sl.Method(fn))

	c, err := r.ctxman.get(ctx)
	if err != nil {
		return err
	}

	mbox, err := c.Select(mailbox, false)
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	log.Debug("mailbox", slog.Any("mailbox", mbox))

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(num)

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := c.Store(seqSet, item, flags, nil); err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	if err := c.Expunge(nil); err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}

	return nil
}
