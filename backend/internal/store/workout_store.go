package store

import "github.com/kevsommer/runplanner/internal/model"

type WorkoutStore interface {
	Create(workout *model.Workout) error
	GetByID(id model.WorkoutID) (*model.Workout, error)
	GetByPlanID(planID model.TrainingPlanID) ([]*model.Workout, error)
	Update(workout *model.Workout) error
	Delete(id model.WorkoutID) error
}
