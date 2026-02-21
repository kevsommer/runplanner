package mem

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

type memUserStore struct {
	mu      sync.RWMutex
	byID    map[model.UserID]*model.User
	byEmail map[string]model.UserID
}

func NewMemUserStore() store.UserStore {
	return &memUserStore{
		byID:    make(map[model.UserID]*model.User),
		byEmail: make(map[string]model.UserID),
	}
}

func (s *memUserStore) CreateUser(email string, passwordHash []byte) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byEmail[email]; exists {
		return nil, store.ErrEmailTaken
	}
	id := model.UserID(newID())
	u := &model.User{ID: id, Email: email, PasswordHash: passwordHash, CreatedAt: time.Now().UTC()}
	s.byID[id] = u
	s.byEmail[email] = id
	return u, nil
}

func (s *memUserStore) GetUserByEmail(email string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byEmail[email]
	if !ok {
		return nil, store.ErrNotFound
	}
	return s.byID[id], nil
}

func (s *memUserStore) GetUserByID(id model.UserID) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byID[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return u, nil
}

func (s *memUserStore) SetActivePlan(userID model.UserID, planID *model.TrainingPlanID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.byID[userID]
	if !ok {
		return store.ErrNotFound
	}
	u.ActivePlanID = planID
	return nil
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
