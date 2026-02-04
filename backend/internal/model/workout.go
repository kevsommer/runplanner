package model

import "time"

type WorkoutID string

type Workout struct {
	ID          WorkoutID      `json:"id"`
	PlanID      TrainingPlanID `json:"planId"`
	RunType     string         `json:"runType"` // e.g., "easy run", "intervals", "long run"
	Day         time.Time      `json:"day"`
	Description string         `json:"description"`
	Notes       string         `json:"notes"`
	Done        bool           `json:"done"`
	Distance    float64        `json:"distance"` // in kilometers
}
