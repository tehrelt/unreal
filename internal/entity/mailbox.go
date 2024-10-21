package entity

type Mailbox struct {
	Name        string   `json:"name"`
	Attributes  []string `json:"attributes"`
	UnreadCount int      `json:"unreadCount"`
}
