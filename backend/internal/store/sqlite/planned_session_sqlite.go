package sqlite

import (
	"database/sql"

	_ "modernc.org/sqlite" // driver

	"github.com/kevsommer/runplanner/internal/model"
)

type PlannedSessionStore struct {
	db *sql.DB
}

func NewPlannedSessionStore(db *sql.DB) *PlannedSessionStore { return &PlannedSessionStore{db: db} }

func (s *PlannedSessionStore) CreatePlannedSession(session *model.PlannedSession) (*model.PlannedSession, error) {
	query := `
		INSERT INTO planned_sessions (plan_id, name, date, workout_type, target_distance_km, status, description, week_index)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		session.PlanID,
		session.Name,
		session.Date,
		session.WorkoutType,
		session.TargetDistanceKm,
		session.Status,
		session.Description,
		session.WeekIndex,
	)

	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *PlannedSessionStore) GetPlannedSessionsForPlan(planID int64) ([]*model.PlannedSession, error) {
	query := `
		SELECT id, plan_id, name, date, workout_type, target_distance_km, status, description, week_index
		FROM planned_sessions
		WHERE plan_id = ?
		ORDER BY date
	`

	rows, err := s.db.Query(query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.PlannedSession
	for rows.Next() {
		var session model.PlannedSession
		if err := rows.Scan(
			&session.ID,
			&session.PlanID,
			&session.Name,
			&session.Date,
			&session.WorkoutType,
			&session.TargetDistanceKm,
			&session.Status,
			&session.Description,
			&session.WeekIndex,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s *PlannedSessionStore) GetPlannedSessionByID(sessionID int64) (*model.PlannedSession, error) {
	query := `
		SELECT id, plan_id, name, date, workout_type, target_distance_km, status, description, week_index
		FROM planned_sessions
		WHERE id = ?
	`

	row := s.db.QueryRow(query, sessionID)
	var session model.PlannedSession
	if err := row.Scan(
		&session.ID,
		&session.PlanID,
		&session.Name,
		&session.Date,
		&session.WorkoutType,
		&session.TargetDistanceKm,
		&session.Status,
		&session.Description,
		&session.WeekIndex,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &session, nil
}

func (s *PlannedSessionStore) UpdatePlannedSession(session *model.PlannedSession) (*model.PlannedSession, error) {
	query := `
		UPDATE planned_sessions
		SET plan_id = ?, name = ?, date = ?, workout_type = ?, target_distance_km = ?, status = ?, description = ?, week_index = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		session.PlanID,
		session.Name,
		session.Date,
		session.WorkoutType,
		session.TargetDistanceKm,
		session.Status,
		session.Description,
		session.WeekIndex,
		session.ID,
	)

	if err != nil {
		return nil, err
	}
	return session, nil
}
