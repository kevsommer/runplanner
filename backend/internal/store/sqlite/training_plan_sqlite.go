package sqlite

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite" // driver

	"github.com/kevsommer/runplanner/internal/model"
)

type TrainingPlanStore struct {
	db *sql.DB
}

func NewTrainingPlanStore(db *sql.DB) *TrainingPlanStore { return &TrainingPlanStore{db: db} }

func (s *TrainingPlanStore) CreateTrainingPlan(plan *model.TrainingPlan) (*model.TrainingPlan, error) {
	query := `
		INSERT INTO training_plans (user_id, goal, start_date, end_date, activities_per_week, name, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(query,
		plan.UserID,
		plan.Goal,
		plan.StartDate,
		plan.EndDate,
		plan.ActivitiesPerWeek,
		plan.Name,
		plan.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *TrainingPlanStore) GetTrainingPlansForUser(userID model.UserID) ([]*model.TrainingPlan, error) {
	query := `
		SELECT id, goal, start_date, end_date, activities_per_week, name, created_at
		FROM training_plans
		WHERE user_id = ?
		ORDER BY start_date
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*model.TrainingPlan
	for rows.Next() {
		var plan model.TrainingPlan
		if err := rows.Scan(
			&plan.ID,
			&plan.Goal,
			&plan.StartDate,
			&plan.EndDate,
			&plan.ActivitiesPerWeek,
			&plan.Name,
			&plan.CreatedAt,
		); err != nil {
			return nil, err
		}
		plans = append(plans, &plan)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func (s *TrainingPlanStore) GetTrainingPlanByID(planID string) (*model.TrainingPlan, error) {
	query := `SELECT id, goal, start_date, end_date, activities_per_week, name, created_at
			  FROM training_plans WHERE id = ?`

	row := s.db.QueryRow(query, planID)

	var plan model.TrainingPlan
	err := row.Scan(
		&plan.ID,
		&plan.Goal,
		&plan.StartDate,
		&plan.EndDate,
		&plan.ActivitiesPerWeek,
		&plan.Name,
		&plan.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &plan, nil
}
