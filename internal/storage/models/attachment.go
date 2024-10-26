package models

import "io"

type Attachment struct {
	R           io.Reader
	ContentType string
	Encrypted   string
}
