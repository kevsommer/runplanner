package mem

import (
	"sync"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

type memWorkoutStore struct {
	mu   sync.RWMutex
	byID map[model.WorkoutID]*model.Workout
}

func NewMemWorkoutStore() store.WorkoutStore {
	return &memWorkoutStore{
		byID: make(map[model.WorkoutID]*model.Workout),
	}
}

func (s *memWorkoutStore) Create(workout *model.Workout) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[workout.ID] = workout
	return nil
}

func (s *memWorkoutStore) GetByID(id model.WorkoutID) (*model.Workout, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.byID[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return w, nil
}

func (s *memWorkoutStore) GetByPlanID(planID model.TrainingPlanID) ([]*model.Workout, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var workouts []*model.Workout
	for _, w := range s.byID {
		if w.PlanID == planID {
			workouts = append(workouts, w)
		}
	}
	// Sort by day ascending
	for i := 0; i < len(workouts); i++ {
		for j := i + 1; j < len(workouts); j++ {
			if workouts[j].Day.Before(workouts[i].Day) {
				workouts[i], workouts[j] = workouts[j], workouts[i]
			}
		}
	}
	return workouts, nil
}

func (s *memWorkoutStore) Update(workout *model.Workout) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[workout.ID]; !ok {
		return store.ErrNotFound
	}
	s.byID[workout.ID] = workout
	return nil
}

func (s *memWorkoutStore) Delete(id model.WorkoutID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.byID, id)
	return nil
}
