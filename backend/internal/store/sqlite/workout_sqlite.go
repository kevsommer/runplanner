package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

type WorkoutStore struct {
	db *sql.DB
}

func NewWorkoutStore(db *sql.DB) *WorkoutStore {
	return &WorkoutStore{db: db}
}

func (s *WorkoutStore) Create(workout *model.Workout) error {
	_, err := s.db.Exec(
		`INSERT INTO workouts (id, plan_id, runType, day, description, notes, done, distance) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		workout.ID, workout.PlanID, workout.RunType, workout.Day.Format(dateFormat), workout.Description, workout.Notes, workout.Done, workout.Distance,
	)
	return err
}

func (s *WorkoutStore) CreateBatch(workouts []*model.Workout) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, w := range workouts {
		_, err := tx.Exec(
			`INSERT INTO workouts (id, plan_id, runType, day, description, notes, done, distance) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			w.ID, w.PlanID, w.RunType, w.Day.Format(dateFormat), w.Description, w.Notes, w.Done, w.Distance,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *WorkoutStore) GetByID(id model.WorkoutID) (*model.Workout, error) {
	row := s.db.QueryRow(
		`SELECT id, plan_id, runType, day, description, notes, done, distance FROM workouts WHERE id = ?`,
		id,
	)
	return scanWorkout(row)
}

func (s *WorkoutStore) GetByPlanID(planID model.TrainingPlanID) ([]*model.Workout, error) {
	rows, err := s.db.Query(
		`SELECT id, plan_id, runType, day, description, notes, done, distance FROM workouts WHERE plan_id = ?`,
		planID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var workouts []*model.Workout
	for rows.Next() {
		workout, err := scanWorkoutFromRows(rows)
		if err != nil {
			return nil, err
		}
		workouts = append(workouts, workout)
	}
	return workouts, rows.Err()
}

func (s *WorkoutStore) Update(workout *model.Workout) error {
	_, err := s.db.Exec(
		`UPDATE workouts SET runType = ?, day = ?, description = ?, notes = ?, done = ?, distance = ? WHERE id = ?`,
		workout.RunType, workout.Day.Format(dateFormat), workout.Description, workout.Notes, workout.Done, workout.Distance, workout.ID,
	)
	return err
}

func (s *WorkoutStore) Delete(id model.WorkoutID) error {
	_, err := s.db.Exec(`DELETE FROM workouts WHERE id = ?`, id)
	return err
}

func scanWorkout(row *sql.Row) (*model.Workout, error) {
	var id, pid, runType, dayStr, description, notes string
	var distance float64
	var done bool
	if err := row.Scan(&id, &pid, &runType, &dayStr, &description, &notes, &done, &distance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	day, _ := time.Parse(dateFormat, dayStr)
	return &model.Workout{
		ID:          model.WorkoutID(id),
		PlanID:      model.TrainingPlanID(pid),
		RunType:     runType,
		Day:         day,
		Description: description,
		Notes:       notes,
		Done:        done,
		Distance:    distance,
	}, nil
}

func scanWorkoutFromRows(rows *sql.Rows) (*model.Workout, error) {
	var id, pid, runType, dayStr, description, notes string
	var distance float64
	var done bool
	if err := rows.Scan(&id, &pid, &runType, &dayStr, &description, &notes, &done, &distance); err != nil {
		return nil, err
	}

	day, _ := time.Parse(dateFormat, dayStr)
	return &model.Workout{
		ID:          model.WorkoutID(id),
		PlanID:      model.TrainingPlanID(pid),
		RunType:     runType,
		Day:         day,
		Description: description,
		Notes:       notes,
		Done:        done,
		Distance:    distance,
	}, nil
}
