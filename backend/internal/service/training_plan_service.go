package service

import (
	"errors"
	"time"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

type TrainingPlanService struct {
	plans store.TrainingPlanStore
}

func NewTrainingPlanService(plans store.TrainingPlanStore) *TrainingPlanService {
	return &TrainingPlanService{plans: plans}
}

var (
	errInvalidGoal = errors.New("invalid goal")
)

func (s *TrainingPlanService) Create(userId string, goal string, startDate time.Time, numberOfWeeks, activitiesPerWeek int, name string) (*model.TrainingPlan, error) {
	if !isGoal(goal) {
		return nil, errInvalidGoal
	}

	newTrainingPlan := &model.TrainingPlan{
		UserID:            userId,
		Goal:              goal,
		StartDate:         startDate,
		EndDate:           calculateEndDate(startDate, numberOfWeeks),
		ActivitiesPerWeek: activitiesPerWeek,
		CreatedAt:         time.Now(),
		Name:              &name,
	}

	tp, err := s.plans.CreateTrainingPlan(newTrainingPlan)
	if err != nil {
		return nil, err
	}
	return tp, nil
}

func (s *TrainingPlanService) GetPlansForUser(userId string) ([]*model.TrainingPlan, error) {
	return s.plans.GetTrainingPlansForUser(userId)
}

func (s *TrainingPlanService) GetPlanByID(planID string) (*model.TrainingPlan, error) {
	return s.plans.GetTrainingPlanByID(planID)
}

func calculateEndDate(startDate time.Time, numberOfWeeks int) time.Time {
	return startDate.AddDate(0, 0, numberOfWeeks*7)
}

func isGoal(s string) bool {
	validGoals := map[string]bool{
		"5K":       true,
		"10K":      true,
		"HALF":     true,
		"MARATHON": true,
	}
	return validGoals[s]
}
