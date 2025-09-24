package model

import (
	"time"
)

type UserID string

type User struct {
	ID           UserID    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type PublicUser struct {
	ID    UserID `json:"id"`
	Email string `json:"email"`
}

func (u *User) Public() PublicUser { return PublicUser{ID: u.ID, Email: u.Email} }
