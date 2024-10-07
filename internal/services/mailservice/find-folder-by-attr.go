package mailservice

import (
	"context"
	"fmt"
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func (ms *MailService) findFolderByAttr(_ context.Context, c *client.Client, attribute string) (string, error) {
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	var folder string

	for m := range mailboxes {
		log.Println("Found mailbox:", m.Name, m.Attributes)
		for _, attr := range m.Attributes {
			if attr == attribute {
				folder = m.Name
				log.Println("Sent folder found:", folder)

				return folder, nil
			}
		}
	}

	if err := <-done; err != nil {
		return "", err
	}

	if folder == "" {
		return "", fmt.Errorf("sent folder not found")
	}

	return folder, nil
}
