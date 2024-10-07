package entity

import "strings"

type MailboxName string

const (
	normdelim = ">"
	delim     = "/"
)

func NewMailboxName(name string) MailboxName {
	return MailboxName(MailboxName(name).String())
}

func (m MailboxName) String() string {
	return strings.ReplaceAll(string(m), delim, normdelim)
}

func (m MailboxName) Normalized() string {
	return strings.ReplaceAll(string(m), normdelim, delim)
}

type Mailbox struct {
	Name        MailboxName `json:"name"`
	Attributes  []string    `json:"attributes"`
	UnreadCount int         `json:"unreadCount"`
}
