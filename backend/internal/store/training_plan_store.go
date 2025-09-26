package store

import "github.com/kevsommer/runplanner/internal/model"

type TrainingPlanStore interface {
	CreateTrainingPlan(plan *model.TrainingPlan) (*model.TrainingPlan, error)
	GetTrainingPlansForUser(userID string) ([]*model.TrainingPlan, error)
	GetTrainingPlanByID(planID string) (*model.TrainingPlan, error)
}
