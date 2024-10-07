package entity

import (
	"time"
)

type AddressInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
type Attachment struct {
	ContentId   string `json:"contentId"`
	ContentType string `json:"contentType"`
	Filename    string `json:"filename"`
}

type Message struct {
	Id       uint32        `json:"id"`
	To       []AddressInfo `json:"to"`
	From     AddressInfo   `json:"from"`
	Subject  string        `json:"subject"`
	SentDate time.Time     `json:"sentDate"`
	IsRead   bool          `json:"isRead"`
}

type MessageWithBody struct {
	Message
	Body        string       `json:"body"`
	Attachments []Attachment `json:"attachments"`
	Embedded    []Attachment `json:"embedded"`
}
