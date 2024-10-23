package models

import "io"

type File struct {
	Filename    string
	ContentType string
	io.Reader
}

func NewFile(rd io.Reader, name string, contentType string) *File {
	return &File{
		Reader:      rd,
		Filename:    name,
		ContentType: contentType,
	}
}
