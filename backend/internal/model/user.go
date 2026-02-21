package model

import (
	"time"
)

type UserID string

type User struct {
	ID           UserID          `json:"id"`
	Email        string          `json:"email"`
	PasswordHash []byte          `json:"-"`
	CreatedAt    time.Time       `json:"createdAt"`
	ActivePlanID *TrainingPlanID `json:"activePlanId,omitempty"`
}

type PublicUser struct {
	ID           UserID          `json:"id"`
	Email        string          `json:"email"`
	ActivePlanID *TrainingPlanID `json:"activePlanId,omitempty"`
}

func (u *User) Public() PublicUser {
	return PublicUser{ID: u.ID, Email: u.Email, ActivePlanID: u.ActivePlanID}
}
