package dto

import "github.com/tehrelt/unreal/internal/entity"

type LoginDto struct {
	Email    string            `json:"email"`
	Password string            `json:"password"`
	Imap     entity.Connection `json:"imap"`
	Smtp     entity.Connection `json:"smtp"`
}
