package storage

import "errors"

var (
	ErrSessionNotFound = errors.New("session not found")

	ErrUserNotFound      = errors.New("users not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)
