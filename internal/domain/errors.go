package domain

import "errors"

var (
	ErrUserNotInContext = errors.New("user not in context")
	ErrMailboxNotFound  = errors.New("mailbox not found")
)
