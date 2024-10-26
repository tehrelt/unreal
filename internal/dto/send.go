package dto

import (
	"fmt"
	"io"
	"mime"
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

	name := mime.QEncoding.Encode("UTF-8", *m.Name)

	return fmt.Sprintf("%s <%s>", name, m.Email)
}

type SendMessageDto struct {
	From          *MailRecord
	To            []*MailRecord
	Subject       string
	Body          io.Reader
	Attachments   []*multipart.FileHeader
	DoEncryiption bool
}
