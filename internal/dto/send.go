package dto

import "io"

type SendMessageDto struct {
	To          []string
	Subject     string
	Body        io.Reader
	Attachments []io.Reader
}
