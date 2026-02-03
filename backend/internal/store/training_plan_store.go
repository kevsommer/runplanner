package store

import "github.com/kevsommer/runplanner/internal/model"

type TrainingPlanStore interface {
	Create(plan *model.TrainingPlan) error
	GetByID(id model.TrainingPlanID) (*model.TrainingPlan, error)
}
