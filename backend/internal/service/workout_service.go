package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

var (
	ErrInvalidDistance = errors.New("distance cannot be negative")
	ErrInvalidRunType  = errors.New("invalid run type")
)

type WorkoutService struct {
	workouts store.WorkoutStore
}

func NewWorkoutService(workouts store.WorkoutStore) *WorkoutService {
	return &WorkoutService{workouts: workouts}
}

func (s *WorkoutService) Create(planID model.TrainingPlanID, runType string, day time.Time, description string, distance float64) (*model.Workout, error) {
	if distance < 0 {
		return nil, ErrInvalidDistance
	}

	if runType != "easy_run" && runType != "intervals" && runType != "long_run" && runType != "tempo_run" {
		return nil, ErrInvalidRunType
	}

	workout := &model.Workout{
		ID:          model.WorkoutID(newPlanID()),
		PlanID:      planID,
		RunType:     runType,
		Day:         day,
		Description: description,
		Notes:       "",
		Done:        false,
		Distance:    distance,
	}
	if err := s.workouts.Create(workout); err != nil {
		return nil, err
	}
	return workout, nil
}

func (s *WorkoutService) GetByID(id model.WorkoutID) (*model.Workout, error) {
	return s.workouts.GetByID(id)
}

func (s *WorkoutService) GetByPlanID(planID model.TrainingPlanID) ([]*model.Workout, error) {
	return s.workouts.GetByPlanID(planID)
}

func newWorkoutID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
