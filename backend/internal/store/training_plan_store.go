package store

import "github.com/kevsommer/runplanner/internal/model"

type TrainingPlanStore interface {
	Create(plan *model.TrainingPlan) error
	GetByID(id model.TrainingPlanID) (*model.TrainingPlan, error)
	GetByUserID(userID model.UserID) ([]*model.TrainingPlan, error)
	Update(plan *model.TrainingPlan) error
	Delete(id model.TrainingPlanID) error
}
