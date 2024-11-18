package models

import (
	"io"
	"mime/multipart"
	"time"

	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
)

type Message struct {
	Id          string
	SeqNum      uint32
	From        entity.AddressInfo
	To          []entity.AddressInfo
	Subject     string
	Body        io.Reader
	Attachments []entity.Attachment
	VaultId     string
	Sign        string
	SentDate    time.Time
	IsRead      bool
}

type SendMessage struct {
	From        *dto.MailRecord
	To          []*dto.MailRecord
	Subject     string
	Body        io.Reader
	Attachments []*multipart.FileHeader
	EncryptKey  string
	Sign        string
}
