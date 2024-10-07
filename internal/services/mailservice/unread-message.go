package mailservice

import (
	"context"
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func (ms *MailService) unreadMessage(ctx context.Context, c *client.Client) (int, error) {

	if c.Mailbox() == nil {
		return 0, fmt.Errorf("no mailbox selected")
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}

	ids, err := c.Search(criteria)
	if err != nil {
		return 0, fmt.Errorf("failed to search mailbox %q: %v", c.Mailbox().Name, err)
	}

	return len(ids), nil
}
