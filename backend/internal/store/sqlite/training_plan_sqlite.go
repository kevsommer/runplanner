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

const trainingPlanColumns = `id, user_id, name, end_date, weeks, start_date, created_at, archived_at`

func (s *TrainingPlanStore) Create(plan *model.TrainingPlan) error {
	_, err := s.db.Exec(
		`INSERT INTO training_plans (`+trainingPlanColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		plan.ID, plan.UserID, plan.Name, plan.EndDate.Format(dateFormat), plan.Weeks, plan.StartDate.Format(dateFormat), plan.CreatedAt, nullTime(plan.ArchivedAt),
	)
	return err
}

func (s *TrainingPlanStore) GetByID(id model.TrainingPlanID) (*model.TrainingPlan, error) {
	row := s.db.QueryRow(
		`SELECT `+trainingPlanColumns+` FROM training_plans WHERE id = ?`,
		id,
	)
	plan, err := scanTrainingPlan(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, store.ErrNotFound
	}
	return plan, err
}

func (s *TrainingPlanStore) GetByUserID(userID model.UserID) ([]*model.TrainingPlan, error) {
	rows, err := s.db.Query(
		`SELECT `+trainingPlanColumns+` FROM training_plans WHERE user_id = ? ORDER BY end_date ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var plans []*model.TrainingPlan
	for rows.Next() {
		plan, err := scanTrainingPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanTrainingPlan(row rowScanner) (*model.TrainingPlan, error) {
	var id, uid, name, endDateStr, startDateStr string
	var weeks int
	var createdAt time.Time
	var archivedAt sql.NullTime
	if err := row.Scan(&id, &uid, &name, &endDateStr, &weeks, &startDateStr, &createdAt, &archivedAt); err != nil {
		return nil, err
	}
	endDate, _ := time.Parse(dateFormat, endDateStr)
	startDate, _ := time.Parse(dateFormat, startDateStr)
	plan := &model.TrainingPlan{
		ID:        model.TrainingPlanID(id),
		UserID:    model.UserID(uid),
		Name:      name,
		EndDate:   endDate,
		Weeks:     weeks,
		StartDate: startDate,
		CreatedAt: createdAt,
	}
	if archivedAt.Valid {
		t := archivedAt.Time
		plan.ArchivedAt = &t
	}
	return plan, nil
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func (s *TrainingPlanStore) Update(plan *model.TrainingPlan) error {
	res, err := s.db.Exec(
		`UPDATE training_plans SET name = ?, end_date = ?, weeks = ?, start_date = ?, archived_at = ? WHERE id = ?`,
		plan.Name, plan.EndDate.Format(dateFormat), plan.Weeks, plan.StartDate.Format(dateFormat), nullTime(plan.ArchivedAt), plan.ID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *TrainingPlanStore) Delete(id model.TrainingPlanID) error {
	res, err := s.db.Exec(`DELETE FROM training_plans WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}
