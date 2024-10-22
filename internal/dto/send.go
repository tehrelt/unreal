package dto

import (
	"fmt"
	"io"
	"mime/multipart"
)

type MailRecord struct {
	Name  *string
	Email string
}

func (m *MailRecord) SetName(name *string) {
	m.Name = name
}

func (m *MailRecord) String() string {
	if m.Name == nil {
		return m.Email
	}

	return fmt.Sprintf("%s <%s>", *m.Name, m.Email)
}

type SendMessageDto struct {
	From        *MailRecord
	To          []*MailRecord
	Subject     string
	Body        io.Reader
	Attachments []*multipart.FileHeader
}
