package store

import "github.com/kevsommer/runplanner/internal/model"

type PlannedSessionStore interface {
	CreatePlannedSession(session *model.PlannedSession) (*model.PlannedSession, error)
	GetPlannedSessionsForPlan(planID int64) ([]*model.PlannedSession, error)
	GetPlannedSessionByID(sessionID int64) (*model.PlannedSession, error)
	UpdatePlannedSession(session *model.PlannedSession) (*model.PlannedSession, error)
}
