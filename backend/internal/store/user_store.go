package store

import "github.com/kevsommer/runplanner/internal/model"

// UserStore defines the persistence boundary.
type UserStore interface {
	CreateUser(email string, passwordHash []byte) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserByID(id model.UserID) (*model.User, error)
}

// Domain errors for portability.
var (
	ErrEmailTaken = Err("email already registered")
	ErrNotFound   = Err("not found")
)

type Err string

func (e Err) Error() string { return string(e) }
