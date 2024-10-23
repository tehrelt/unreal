package entity

import (
	"mime/multipart"
)

type Credentials struct {
	Email    string
	Password string
}

type Connection struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type User struct {
	Email   string
	Name    *string
	Picture *string
}

type UpdateUser struct {
	Email   string
	Name    *string
	Picture *multipart.FileHeader
}
