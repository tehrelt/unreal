package models

import "time"

type UserBase struct {
	Email string
}

type CreateUser struct {
	UserBase
}

type UpdateUser struct {
	UserBase
	Name           *string
	ProfilePicture *string
}

type User struct {
	UserBase
	Name           *string
	ProfilePicture *string
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}
