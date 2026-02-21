package service

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

type AuthService struct {
	users store.UserStore
}

func NewAuthService(users store.UserStore) *AuthService { return &AuthService{users: users} }

var (
	errInvalidEmail   = errors.New("invalid email")
	errWeakPassword   = errors.New("password must be at least 8 chars")
	errBadCredentials = errors.New("invalid email or password")
)

func (s *AuthService) Register(email, password string) (*model.User, error) {
	if !isEmail(email) {
		return nil, errInvalidEmail
	}
	if len(password) < 8 {
		return nil, errWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u, err := s.users.CreateUser(email, hash)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *AuthService) Login(email, password string) (*model.User, error) {
	u, err := s.users.GetUserByEmail(email)
	if err != nil {
		return nil, errBadCredentials
	}
	if bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)) != nil {
		return nil, errBadCredentials
	}
	return u, nil
}

func (s *AuthService) GetUser(id model.UserID) (*model.User, error) {
	return s.users.GetUserByID(id)
}

func (s *AuthService) SetActivePlan(userID model.UserID, planID *model.TrainingPlanID) error {
	return s.users.SetActivePlan(userID, planID)
}

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func isEmail(s string) bool { return emailRe.MatchString(s) }
