package models

import "time"

type CreateUser struct {
	Email string
}

type User struct {
	Id             string
	Email          string
	Name           *string
	ProfilePicture *string
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}
