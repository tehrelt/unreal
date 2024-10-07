package dto

import (
	"io"
	"mime/multipart"
)

type SendMessageDto struct {
	To          []string
	Subject     string
	Body        io.Reader
	Attachments []*multipart.FileHeader
}
