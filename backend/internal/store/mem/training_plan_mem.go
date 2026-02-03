package mem

import (
	"sync"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

type memTrainingPlanStore struct {
	mu   sync.RWMutex
	byID map[model.TrainingPlanID]*model.TrainingPlan
}

func NewMemTrainingPlanStore() store.TrainingPlanStore {
	return &memTrainingPlanStore{
		byID: make(map[model.TrainingPlanID]*model.TrainingPlan),
	}
}

func (s *memTrainingPlanStore) Create(plan *model.TrainingPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[plan.ID] = plan
	return nil
}

func (s *memTrainingPlanStore) GetByID(id model.TrainingPlanID) (*model.TrainingPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.byID[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return p, nil
}

func (s *memTrainingPlanStore) GetByUserID(userID model.UserID) ([]*model.TrainingPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var plans []*model.TrainingPlan
	for _, p := range s.byID {
		if p.UserID == userID {
			plans = append(plans, p)
		}
	}
	// Sort by end_date ascending
	for i := 0; i < len(plans); i++ {
		for j := i + 1; j < len(plans); j++ {
			if plans[j].EndDate.Before(plans[i].EndDate) {
				plans[i], plans[j] = plans[j], plans[i]
			}
		}
	}
	return plans, nil
}

func (s *memTrainingPlanStore) Update(plan *model.TrainingPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[plan.ID]; !ok {
		return store.ErrNotFound
	}
	s.byID[plan.ID] = plan
	return nil
}

func (s *memTrainingPlanStore) Delete(id model.TrainingPlanID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.byID, id)
	return nil
}
