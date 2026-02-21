package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

var (
	ErrInvalidDistance             = errors.New("distance cannot be negative")
	ErrInvalidRunType              = errors.New("invalid run type")
	ErrInvalidStatus               = errors.New("invalid status")
	ErrStrengthTrainingNonZeroDist = errors.New("strength training must have a distance of 0km")
	ErrInvalidRaceGoal             = errors.New("invalid race goal: must be one of 5k, 10k, halfmarathon, marathon")
)

var RaceGoalDistances = map[string]float64{
	"5k":           5,
	"10k":          10,
	"halfmarathon": 21,
	"marathon":     42,
}

var raceGoalLabels = map[string]string{
	"5k":           "5K",
	"10k":          "10K",
	"halfmarathon": "Half Marathon",
	"marathon":     "Marathon",
}

type BatchValidationError struct {
	Index   int
	Message string
}

func (e *BatchValidationError) Error() string {
	return fmt.Sprintf("workout[%d]: %s", e.Index, e.Message)
}

type BulkWorkoutInput struct {
	RunType     string
	Week        int
	DayOfWeek   int // 1=Monday, 7=Sunday
	Description string
	Distance    float64
}

type WorkoutService struct {
	workouts store.WorkoutStore
}

func NewWorkoutService(workouts store.WorkoutStore) *WorkoutService {
	return &WorkoutService{workouts: workouts}
}

func isValidRunType(runType string) bool {
	validRunTypes := map[string]bool{
		"easy_run": true, 
		"intervals": true, 
		"long_run": true,
		"tempo_run": true,
		"strength_training": true,
		"race": true,
	}
	return validRunTypes[runType]
}

func (s *WorkoutService) Create(planID model.TrainingPlanID, runType string, day time.Time, description string, distance float64) (*model.Workout, error) {
	if distance < 0 {
		return nil, ErrInvalidDistance
	}

	if !isValidRunType(runType) {
		return nil, ErrInvalidRunType
	}

	if runType == "strength_training" && distance != 0 {
		return nil, ErrStrengthTrainingNonZeroDist
	}

	workout := &model.Workout{
		ID:          model.WorkoutID(newPlanID()),
		PlanID:      planID,
		RunType:     runType,
		Day:         day,
		Description: description,
		Notes:       "",
		Status:      "pending",
		Distance:    distance,
	}
	if err := s.workouts.Create(workout); err != nil {
		return nil, err
	}
	return workout, nil
}

func (s *WorkoutService) CreateBatch(plan *model.TrainingPlan, items []BulkWorkoutInput) ([]*model.Workout, error) {
	workouts := make([]*model.Workout, 0, len(items))
	for i, item := range items {
		if !isValidRunType(item.RunType) {
			return nil, &BatchValidationError{Index: i, Message: "invalid run type"}
		}
		if item.Distance < 0 {
			return nil, &BatchValidationError{Index: i, Message: "distance cannot be negative"}
		}
		if item.RunType == "strength_training" && item.Distance != 0 {
			return nil, &BatchValidationError{Index: i, Message: "strength training must have a distance of 0km"}
		}
		if item.Week < 1 || item.Week > plan.Weeks {
			return nil, &BatchValidationError{Index: i, Message: fmt.Sprintf("week must be between 1 and %d", plan.Weeks)}
		}
		if item.DayOfWeek < 1 || item.DayOfWeek > 7 {
			return nil, &BatchValidationError{Index: i, Message: "dayOfWeek must be between 1 (Monday) and 7 (Sunday)"}
		}
		day := plan.StartDate.AddDate(0, 0, (item.Week-1)*7+(item.DayOfWeek-1))
		workouts = append(workouts, &model.Workout{
			ID:          model.WorkoutID(newWorkoutID()),
			PlanID:      plan.ID,
			RunType:     item.RunType,
			Day:         day,
			Description: item.Description,
			Status:      "pending",
			Distance:    item.Distance,
		})
	}

	if err := s.workouts.CreateBatch(workouts); err != nil {
		return nil, err
	}
	return workouts, nil
}

func (s *WorkoutService) GetByID(id model.WorkoutID) (*model.Workout, error) {
	return s.workouts.GetByID(id)
}

func (s *WorkoutService) GetByPlanID(planID model.TrainingPlanID) ([]*model.Workout, error) {
	return s.workouts.GetByPlanID(planID)
}

func (s *WorkoutService) Update(workout *model.Workout) error {
	if workout.Distance < 0 {
		return ErrInvalidDistance
	}

	if !isValidRunType(workout.RunType) {
		return ErrInvalidRunType
	}

	if workout.RunType == "strength_training" && workout.Distance != 0 {
		return ErrStrengthTrainingNonZeroDist
	}

	if workout.Status != "pending" && workout.Status != "completed" && workout.Status != "skipped" {
		return ErrInvalidStatus
	}

	return s.workouts.Update(workout)
}

func (s *WorkoutService) CreateRaceWorkout(plan *model.TrainingPlan, raceGoal string) (*model.Workout, error) {
	distance, ok := RaceGoalDistances[raceGoal]
	if !ok {
		return nil, ErrInvalidRaceGoal
	}
	desc := "Race Day - " + raceGoalLabels[raceGoal]
	return s.Create(plan.ID, "race", plan.EndDate, desc, distance)
}

func (s *WorkoutService) Delete(id model.WorkoutID) error {
	return s.workouts.Delete(id)
}

func newWorkoutID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
