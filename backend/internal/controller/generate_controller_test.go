package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/kevsommer/runplanner/internal/ai"
	"github.com/kevsommer/runplanner/internal/service"
	"github.com/kevsommer/runplanner/internal/store/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAIClient struct {
	response string
	err      error
}

func (m *mockAIClient) Complete(_ context.Context, _ ai.CompletionRequest) (string, error) {
	return m.response, m.err
}

func validMockResponse() string {
	return `{"workouts": [
		{"runType": "easy_run", "week": 1, "dayOfWeek": 1, "description": "Easy 5k", "distance": 5.0},
		{"runType": "easy_run", "week": 1, "dayOfWeek": 3, "description": "Easy 6k", "distance": 6.0},
		{"runType": "long_run", "week": 1, "dayOfWeek": 6, "description": "Long run", "distance": 12.0}
	]}`
}

func setupGenerateTestRouter(t *testing.T, mock ai.Client) (*gin.Engine, *service.AuthService) {
	gin.SetMode(gin.TestMode)
	userStore := mem.NewMemUserStore()
	planStore := mem.NewMemTrainingPlanStore()
	workoutStore := mem.NewMemWorkoutStore()
	authSvc := service.NewAuthService(userStore)
	planSvc := service.NewTrainingPlanService(planStore)
	workoutSvc := service.NewWorkoutService(workoutStore)
	genSvc := service.NewGenerateService(mock, planSvc, workoutSvc)

	r := gin.New()
	storeCookie := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("rp.sid", storeCookie))

	api := r.Group("/api")
	RegisterAuthRoutes(api, authSvc)
	RegisterTrainingPlanRoutes(api, planSvc, workoutSvc, genSvc)

	return r, authSvc
}

func loginForGenerate(t *testing.T, r *gin.Engine, email, password string) []*http.Cookie {
	body := map[string]string{"email": email, "password": password}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	return w.Result().Cookies()
}

func TestGenerateController_PostGenerate(t *testing.T) {
	t.Run("generates plan and returns 201", func(t *testing.T) {
		mock := &mockAIClient{response: validMockResponse()}
		r, authSvc := setupGenerateTestRouter(t, mock)
		_, err := authSvc.Register("gen@example.com", "password123")
		require.NoError(t, err)
		cookies := loginForGenerate(t, r, "gen@example.com", "password123")

		body := map[string]interface{}{
			"name":          "My AI Plan",
			"endDate":       "2025-09-15",
			"weeks":         8,
			"baseKmPerWeek": 30.0,
			"runsPerWeek":   3,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/generate", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		plan, ok := resp["plan"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "My AI Plan", plan["name"])
		assert.Equal(t, float64(8), plan["weeks"])

		workouts, ok := resp["workouts"].([]interface{})
		require.True(t, ok)
		assert.Len(t, workouts, 3)
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		mock := &mockAIClient{response: validMockResponse()}
		r, _ := setupGenerateTestRouter(t, mock)

		body := map[string]interface{}{
			"name":          "Plan",
			"endDate":       "2025-09-15",
			"weeks":         8,
			"baseKmPerWeek": 30.0,
			"runsPerWeek":   3,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/generate", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing required fields returns 400", func(t *testing.T) {
		mock := &mockAIClient{response: validMockResponse()}
		r, authSvc := setupGenerateTestRouter(t, mock)
		_, _ = authSvc.Register("gen2@example.com", "password123")
		cookies := loginForGenerate(t, r, "gen2@example.com", "password123")

		body := map[string]interface{}{
			"name": "Plan",
			// missing other fields
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/generate", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid endDate returns 400", func(t *testing.T) {
		mock := &mockAIClient{response: validMockResponse()}
		r, authSvc := setupGenerateTestRouter(t, mock)
		_, _ = authSvc.Register("gen3@example.com", "password123")
		cookies := loginForGenerate(t, r, "gen3@example.com", "password123")

		body := map[string]interface{}{
			"name":          "Plan",
			"endDate":       "not-a-date",
			"weeks":         8,
			"baseKmPerWeek": 30.0,
			"runsPerWeek":   3,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/generate", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["error"], "endDate must be YYYY-MM-DD")
	})

	t.Run("validation error returns 400", func(t *testing.T) {
		mock := &mockAIClient{response: validMockResponse()}
		r, authSvc := setupGenerateTestRouter(t, mock)
		_, _ = authSvc.Register("gen4@example.com", "password123")
		cookies := loginForGenerate(t, r, "gen4@example.com", "password123")

		body := map[string]interface{}{
			"name":          "Plan",
			"endDate":       "2025-09-15",
			"weeks":         3, // less than 6
			"baseKmPerWeek": 30.0,
			"runsPerWeek":   3,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/generate", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["error"], "weeks must be at least 6")
	})

	t.Run("AI client nil returns 400", func(t *testing.T) {
		r, authSvc := setupGenerateTestRouter(t, nil)
		_, _ = authSvc.Register("gen5@example.com", "password123")
		cookies := loginForGenerate(t, r, "gen5@example.com", "password123")

		body := map[string]interface{}{
			"name":          "Plan",
			"endDate":       "2025-09-15",
			"weeks":         8,
			"baseKmPerWeek": 30.0,
			"runsPerWeek":   3,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/generate", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["error"], "not configured")
	})

	t.Run("AI failure returns 502", func(t *testing.T) {
		mock := &mockAIClient{err: errors.New("rate limited")}
		r, authSvc := setupGenerateTestRouter(t, mock)
		_, _ = authSvc.Register("gen6@example.com", "password123")
		cookies := loginForGenerate(t, r, "gen6@example.com", "password123")

		body := map[string]interface{}{
			"name":          "Plan",
			"endDate":       "2025-09-15",
			"weeks":         8,
			"baseKmPerWeek": 30.0,
			"runsPerWeek":   3,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/generate", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("runsPerWeek out of range returns 400", func(t *testing.T) {
		mock := &mockAIClient{response: validMockResponse()}
		r, authSvc := setupGenerateTestRouter(t, mock)
		_, _ = authSvc.Register("gen7@example.com", "password123")
		cookies := loginForGenerate(t, r, "gen7@example.com", "password123")

		body := map[string]interface{}{
			"name":          "Plan",
			"endDate":       "2025-09-15",
			"weeks":         8,
			"baseKmPerWeek": 30.0,
			"runsPerWeek":   9,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/generate", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
