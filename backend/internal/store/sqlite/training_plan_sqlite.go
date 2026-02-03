package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

type TrainingPlanStore struct {
	db *sql.DB
}

func NewTrainingPlanStore(db *sql.DB) *TrainingPlanStore {
	return &TrainingPlanStore{db: db}
}

const dateFormat = "2006-01-02"

func (s *TrainingPlanStore) Create(plan *model.TrainingPlan) error {
	_, err := s.db.Exec(
		`INSERT INTO training_plans (id, user_id, name, end_date, weeks, start_date, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		plan.ID, plan.UserID, plan.Name, plan.EndDate.Format(dateFormat), plan.Weeks, plan.StartDate.Format(dateFormat), plan.CreatedAt,
	)
	return err
}

func (s *TrainingPlanStore) GetByID(id model.TrainingPlanID) (*model.TrainingPlan, error) {
	row := s.db.QueryRow(
		`SELECT id, user_id, name, end_date, weeks, start_date, created_at FROM training_plans WHERE id = ?`,
		id,
	)
	return scanTrainingPlan(row)
}

func (s *TrainingPlanStore) GetByUserID(userID model.UserID) ([]*model.TrainingPlan, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, name, end_date, weeks, start_date, created_at FROM training_plans WHERE user_id = ? ORDER BY end_date ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var plans []*model.TrainingPlan
	for rows.Next() {
		plan, err := scanTrainingPlanFromRows(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func scanTrainingPlan(row *sql.Row) (*model.TrainingPlan, error) {
	var id, uid, name, endDateStr, startDateStr string
	var weeks int
	var createdAt time.Time
	if err := row.Scan(&id, &uid, &name, &endDateStr, &weeks, &startDateStr, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	endDate, _ := time.Parse(dateFormat, endDateStr)
	startDate, _ := time.Parse(dateFormat, startDateStr)
	return &model.TrainingPlan{
		ID:        model.TrainingPlanID(id),
		UserID:    model.UserID(uid),
		Name:      name,
		EndDate:   endDate,
		Weeks:     weeks,
		StartDate: startDate,
		CreatedAt: createdAt,
	}, nil
}

func scanTrainingPlanFromRows(rows *sql.Rows) (*model.TrainingPlan, error) {
	var id, uid, name, endDateStr, startDateStr string
	var weeks int
	var createdAt time.Time
	if err := rows.Scan(&id, &uid, &name, &endDateStr, &weeks, &startDateStr, &createdAt); err != nil {
		return nil, err
	}
	endDate, _ := time.Parse(dateFormat, endDateStr)
	startDate, _ := time.Parse(dateFormat, startDateStr)
	return &model.TrainingPlan{
		ID:        model.TrainingPlanID(id),
		UserID:    model.UserID(uid),
		Name:      name,
		EndDate:   endDate,
		Weeks:     weeks,
		StartDate: startDate,
		CreatedAt: createdAt,
	}, nil
}
