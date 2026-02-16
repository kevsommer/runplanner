package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/kevsommer/runplanner/internal/service"
	"github.com/kevsommer/runplanner/internal/store/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthTestRouter(t *testing.T) (*gin.Engine, *service.AuthService) {
	gin.SetMode(gin.TestMode)
	userStore := mem.NewMemUserStore()
	authSvc := service.NewAuthService(userStore)

	r := gin.New()
	storeCookie := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("rp.sid", storeCookie))

	api := r.Group("/api")
	RegisterAuthRoutes(api, authSvc)

	return r, authSvc
}

func TestAuthController_Register(t *testing.T) {
	r, _ := setupAuthTestRouter(t)

	t.Run("valid registration returns 201 and user", func(t *testing.T) {
		body := map[string]string{"email": "test@example.com", "password": "password123"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		user, ok := resp["user"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "test@example.com", user["email"])
		assert.NotEmpty(t, user["id"])
		// Session cookie should be set (auto-login)
		assert.Contains(t, w.Header().Get("Set-Cookie"), "rp.sid")
	})

	t.Run("register auto-logs in user", func(t *testing.T) {
		body := map[string]string{"email": "autologin@example.com", "password": "password123"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		// Use the session cookie from registration to call /auth/me
		cookies := w.Result().Cookies()
		meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		for _, c := range cookies {
			meReq.AddCookie(c)
		}
		meW := httptest.NewRecorder()
		r.ServeHTTP(meW, meReq)

		assert.Equal(t, http.StatusOK, meW.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(meW.Body.Bytes(), &resp))
		user, ok := resp["user"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "autologin@example.com", user["email"])
	})

	t.Run("missing email returns 400", func(t *testing.T) {
		body := map[string]string{"password": "password123"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "email and password required", resp["error"])
	})

	t.Run("missing password returns 400", func(t *testing.T) {
		body := map[string]string{"email": "test@example.com"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "email and password required", resp["error"])
	})

	t.Run("invalid email returns 400", func(t *testing.T) {
		body := map[string]string{"email": "notanemail", "password": "password123"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("weak password returns 400", func(t *testing.T) {
		body := map[string]string{"email": "test@example.com", "password": "short"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("duplicate email returns 409", func(t *testing.T) {
		body := map[string]string{"email": "dup@example.com", "password": "password123"}
		bodyBytes, _ := json.Marshal(body)

		// First registration succeeds
		req1 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusCreated, w1.Code)

		// Second registration with same email fails
		req2 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusConflict, w2.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp))
		assert.Equal(t, "email already registered", resp["error"])
	})
}

func TestAuthController_Login(t *testing.T) {
	r, authSvc := setupAuthTestRouter(t)

	// Pre-register a user
	u, err := authSvc.Register("login@example.com", "password123")
	require.NoError(t, err)
	require.NotNil(t, u)

	t.Run("valid login returns 200 and user", func(t *testing.T) {
		body := map[string]string{"email": "login@example.com", "password": "password123"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		user, ok := resp["user"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "login@example.com", user["email"])
		assert.Contains(t, w.Header().Get("Set-Cookie"), "rp.sid")
	})

	t.Run("wrong password returns 401", func(t *testing.T) {
		body := map[string]string{"email": "login@example.com", "password": "wrongpassword"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "invalid email or password", resp["error"])
	})

	t.Run("unknown email returns 401", func(t *testing.T) {
		body := map[string]string{"email": "unknown@example.com", "password": "password123"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "invalid email or password", resp["error"])
	})

	t.Run("missing credentials returns 400", func(t *testing.T) {
		body := map[string]string{}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "email and password required", resp["error"])
	})
}

func TestAuthController_Logout(t *testing.T) {
	r, authSvc := setupAuthTestRouter(t)

	u, err := authSvc.Register("logout@example.com", "password123")
	require.NoError(t, err)
	require.NotNil(t, u)

	t.Run("logout returns 204", func(t *testing.T) {
		// Login first to get session
		loginBody := map[string]string{"email": "logout@example.com", "password": "password123"}
		loginBytes, _ := json.Marshal(loginBody)
		loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBytes))
		loginReq.Header.Set("Content-Type", "application/json")
		loginW := httptest.NewRecorder()
		r.ServeHTTP(loginW, loginReq)
		require.Equal(t, http.StatusOK, loginW.Code)

		// Extract session cookie
		cookies := loginW.Result().Cookies()

		// Logout with session
		logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		for _, c := range cookies {
			logoutReq.AddCookie(c)
		}
		logoutW := httptest.NewRecorder()
		r.ServeHTTP(logoutW, logoutReq)

		assert.Equal(t, http.StatusNoContent, logoutW.Code)
	})
}

func TestAuthController_GetMe(t *testing.T) {
	r, authSvc := setupAuthTestRouter(t)

	u, err := authSvc.Register("me@example.com", "password123")
	require.NoError(t, err)
	require.NotNil(t, u)

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "not authenticated", resp["error"])
	})

	t.Run("authenticated returns user", func(t *testing.T) {
		// Login to get session
		loginBody := map[string]string{"email": "me@example.com", "password": "password123"}
		loginBytes, _ := json.Marshal(loginBody)
		loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBytes))
		loginReq.Header.Set("Content-Type", "application/json")
		loginW := httptest.NewRecorder()
		r.ServeHTTP(loginW, loginReq)
		require.Equal(t, http.StatusOK, loginW.Code)

		cookies := loginW.Result().Cookies()

		// Get /me with session
		meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		for _, c := range cookies {
			meReq.AddCookie(c)
		}
		meW := httptest.NewRecorder()
		r.ServeHTTP(meW, meReq)

		assert.Equal(t, http.StatusOK, meW.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(meW.Body.Bytes(), &resp))
		user, ok := resp["user"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "me@example.com", user["email"])
		assert.Equal(t, string(u.ID), user["id"])
	})
}
