package imap

import (
	"context"
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (r *Repository) unread(ctx context.Context) (int, error) {

	fn := "mail.unread"
	log := r.l.With(sl.Method(fn))

	c, err := r.ctxman.get(ctx)
	if err != nil {
		return 0, err
	}

	if c.Mailbox() == nil {
		log.Warn("no mailbox selected for count unread messages")
		return 0, fmt.Errorf("%s: no mailbox selected", fn)
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}

	ids, err := c.Search(criteria)
	if err != nil {
		return 0, fmt.Errorf("failed to search mailbox %q: %v", c.Mailbox().Name, err)
	}

	return len(ids), nil
}
