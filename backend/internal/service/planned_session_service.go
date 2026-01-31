package service

import (
	"time"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

type PlannedSessionService struct {
	sessions store.PlannedSessionStore
}

func NewPlannedSessionService(sessions store.PlannedSessionStore) *PlannedSessionService {
	return &PlannedSessionService{sessions: sessions}
}

func getWeekIndex(date time.Time, planID int64) (int, error) {
	// Placeholder implementation; in a real scenario, this would calculate the week index based on the plan's start date.
	return 0, nil
}

func (s *PlannedSessionService) Create(planID int64, name string, date time.Time, workoutType string, targetDistanceKm float64, description string) (*model.PlannedSession, error) {

	weekIndex, err := getWeekIndex(date, planID)

	if err != nil {
		return nil, err
	}

	newSession := &model.PlannedSession{
		PlanID:           planID,
		Name:             &name,
		Date:             date,
		WorkoutType:      workoutType,
		TargetDistanceKm: &targetDistanceKm,
		Status:           "planned",
		Description:      &description,
		WeekIndex:        weekIndex,
	}

	ps, err := s.sessions.CreatePlannedSession(newSession)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (s *PlannedSessionService) GetSessionsForPlan(planID int64) ([]*model.PlannedSession, error) {
	return s.sessions.GetPlannedSessionsForPlan(planID)
}

func (s *PlannedSessionService) GetSessionByID(sessionID int64) (*model.PlannedSession, error) {
	return s.sessions.GetPlannedSessionByID(sessionID)
}
