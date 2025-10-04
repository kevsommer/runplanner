package model

import (
	"time"
)

type PlannedSessionId string

type PlannedSession struct {
	ID               int64     `db:"id" json:"id"`
	PlanID           int64     `db:"plan_id" json:"planId"`
	Name             *string   `db:"name" json:"name,omitempty"`
	Date             time.Time `db:"date" json:"date"`
	WorkoutType      string    `db:"workout_type" json:"workoutType"`
	TargetDistanceKm *float64  `db:"target_distance_km" json:"targetDistanceKm,omitempty"`
	Status           string    `db:"status" json:"status"`
	Description      *string   `db:"description" json:"description,omitempty"`
	WeekIndex        int       `db:"week_index" json:"weekIndex"`
}
