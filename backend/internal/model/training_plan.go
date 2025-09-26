package model

import (
	"time"
)

type TrainingPlanId string

type TrainingPlan struct {
	ID                int64     `db:"id" json:"id"`
	UserID            string    `db:"user_id" json:"userId"`
	Goal              string    `db:"goal" json:"goal"` // 5K,10K,HALF,MARATHON
	StartDate         time.Time `db:"start_date" json:"startDate"`
	EndDate           time.Time `db:"end_date" json:"endDate"`
	ActivitiesPerWeek int       `db:"activities_per_week" json:"activitiesPerWeek"`
	Name              *string   `db:"name" json:"name,omitempty"`
	CreatedAt         time.Time `db:"created_at" json:"createdAt"`
}
