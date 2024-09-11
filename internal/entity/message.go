package entity

import (
	"bytes"
	"io"
	"strings"
	"time"
)

type AddressInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Message struct {
	Id       uint32        `json:"id"`
	To       []AddressInfo `json:"to"`
	From     AddressInfo   `json:"from"`
	Subject  string        `json:"subject"`
	SentDate time.Time     `json:"sentDate"`
	IsRead   bool          `json:"isRead"`
}

type Body struct {
	ContentType string `json:"contentType"`
	Body        IBody  `json:"body"`
}

type IBody interface {
	Reader() io.Reader
}

type PlainBody string

func (b PlainBody) Reader() io.Reader {
	return strings.NewReader(string(b))
}

type BytesBody []byte

func (b BytesBody) Reader() io.Reader {
	return bytes.NewBuffer(b)
}

type MessageWithBody struct {
	Message
	Content []Body `json:"content"`
}
