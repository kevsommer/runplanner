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
	ErrInvalidWeeks = errors.New("weeks must be at least 1")
	ErrInvalidName  = errors.New("name is required")
)

type TrainingPlanService struct {
	plans store.TrainingPlanStore
}

func NewTrainingPlanService(plans store.TrainingPlanStore) *TrainingPlanService {
	return &TrainingPlanService{plans: plans}
}

func StartDateFor(endDate time.Time, weeks int) time.Time {
	weekday := endDate.Weekday()
	daysSinceMonday := int(weekday) - 1
	if weekday == time.Sunday {
		daysSinceMonday = 6
	}
	mondayOfRaceWeek := endDate.AddDate(0, 0, -daysSinceMonday)
	mondayOfWeek1 := mondayOfRaceWeek.AddDate(0, 0, -(weeks-1)*7)
	return time.Date(mondayOfWeek1.Year(), mondayOfWeek1.Month(), mondayOfWeek1.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *TrainingPlanService) Create(userID model.UserID, name string, endDate time.Time, weeks int) (*model.TrainingPlan, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if weeks < 1 {
		return nil, ErrInvalidWeeks
	}
	startDate := StartDateFor(endDate, weeks)
	plan := &model.TrainingPlan{
		ID:        model.TrainingPlanID(newPlanID()),
		UserID:    userID,
		Name:      name,
		EndDate:   endDate,
		Weeks:     weeks,
		StartDate: startDate,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.plans.Create(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *TrainingPlanService) GetByID(id model.TrainingPlanID) (*model.TrainingPlan, error) {
	return s.plans.GetByID(id)
}

func newPlanID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
