package model

import "time"

type TrainingPlanID string

type TrainingPlan struct {
	ID        TrainingPlanID `json:"id"`
	UserID    UserID         `json:"userId"`
	Name      string         `json:"name"`
	EndDate   time.Time      `json:"endDate"`   // race date
	Weeks     int            `json:"weeks"`     // number of weeks in the plan
	StartDate time.Time      `json:"startDate"` // first Monday of week 1 (calculated)
	CreatedAt time.Time      `json:"createdAt"`
}
