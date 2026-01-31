package service

import (
	"testing"

	"github.com/kevsommer/runplanner/internal/store"
	"github.com/kevsommer/runplanner/internal/store/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func setupTest(t *testing.T) *AuthService {
	return NewAuthService(mem.NewMemUserStore())
}

func TestAuthService_Register(t *testing.T) {
	svc := setupTest(t)

	t.Run("valid registration returns user", func(t *testing.T) {
		user, err := svc.Register("test@example.com", "password123")

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
		assert.NotEmpty(t, user.ID)
		assert.NotEmpty(t, user.PasswordHash)

		err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte("password123"))
		assert.NoError(t, err)
	})

	t.Run("invalid email returns errInvalidEmail", func(t *testing.T) {
		user, err := svc.Register("notanemail", "password123")

		assert.Error(t, err)
		assert.Equal(t, errInvalidEmail, err)
		assert.Nil(t, user)
	})

	t.Run("empty email returns errInvalidEmail", func(t *testing.T) {
		user, err := svc.Register("", "password123")

		assert.Error(t, err)
		assert.Equal(t, errInvalidEmail, err)
		assert.Nil(t, user)
	})

	t.Run("weak password returns errWeakPassword", func(t *testing.T) {
		user, err := svc.Register("test@example.com", "short")

		assert.Error(t, err)
		assert.Equal(t, errWeakPassword, err)
		assert.Nil(t, user)
	})

	t.Run("password exactly 7 chars returns errWeakPassword", func(t *testing.T) {
		user, err := svc.Register("test@example.com", "1234567")

		assert.Error(t, err)
		assert.Equal(t, errWeakPassword, err)
		assert.Nil(t, user)
	})

	t.Run("password exactly 8 chars succeeds", func(t *testing.T) {
		user, err := svc.Register("eight@example.com", "12345678")

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "eight@example.com", user.Email)
	})

	t.Run("duplicate email returns ErrEmailTaken", func(t *testing.T) {
		_, err := svc.Register("dup@example.com", "password123")
		require.NoError(t, err)

		user, err := svc.Register("dup@example.com", "password123")

		assert.Error(t, err)
		assert.Equal(t, store.ErrEmailTaken, err)
		assert.Nil(t, user)
	})
}

func TestAuthService_Login(t *testing.T) {
	svc := setupTest(t)

	// Pre-register a user
	_, err := svc.Register("login@example.com", "password123")
	require.NoError(t, err)

	t.Run("valid credentials returns user", func(t *testing.T) {
		user, err := svc.Login("login@example.com", "password123")

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "login@example.com", user.Email)
		assert.NotEmpty(t, user.ID)
	})

	t.Run("wrong password returns errBadCredentials", func(t *testing.T) {
		user, err := svc.Login("login@example.com", "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, errBadCredentials, err)
		assert.Nil(t, user)
	})

	t.Run("unknown email returns errBadCredentials", func(t *testing.T) {
		user, err := svc.Login("unknown@example.com", "password123")

		assert.Error(t, err)
		assert.Equal(t, errBadCredentials, err)
		assert.Nil(t, user)
	})
}

func TestAuthService_GetUser(t *testing.T) {
	svc := setupTest(t)

	u, err := svc.Register("getuser@example.com", "password123")
	require.NoError(t, err)
	require.NotNil(t, u)

	t.Run("existing user returns user", func(t *testing.T) {
		user, err := svc.GetUser(u.ID)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, u.ID, user.ID)
		assert.Equal(t, "getuser@example.com", user.Email)
	})

	t.Run("non-existent user returns store.ErrNotFound", func(t *testing.T) {
		user, err := svc.GetUser("nonexistent-id-12345")

		assert.Error(t, err)
		assert.Equal(t, store.ErrNotFound, err)
		assert.Nil(t, user)
	})
}
