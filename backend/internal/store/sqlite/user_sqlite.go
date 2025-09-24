package sqlite

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	_ "modernc.org/sqlite" // driver

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore { return &UserStore{db: db} }

func Open(dsn string) (*sql.DB, error) {
	// Example DSN: file:data/runplanner.db?_pragma=busy_timeout(5000)&cache=shared
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite is file-based; one open conn is usually fine.
	db.SetMaxOpenConns(1)
	return db, nil
}

func (s *UserStore) CreateUser(email string, hash []byte) (*model.User, error) {
	id := model.UserID(newID())
	now := time.Now().UTC()

	_, err := s.db.Exec(
		`INSERT INTO users (id, email, password_hash, created_at) VALUES (?, ?, ?, ?)`,
		id, email, hash, now,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, store.ErrEmailTaken
		}
		return nil, err
	}
	return &model.User{ID: id, Email: email, PasswordHash: hash, CreatedAt: now}, nil
}

func (s *UserStore) GetUserByEmail(email string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE email = ?`,
		email,
	)
	u := model.User{}
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) GetUserByID(id model.UserID) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE id = ?`,
		id,
	)
	u := model.User{}
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
