package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/kevsommer/runplanner/internal/service"
	"github.com/kevsommer/runplanner/internal/store/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func setupPlansTestRouter(t *testing.T) (*gin.Engine, *service.AuthService, *service.TrainingPlanService) {
	gin.SetMode(gin.TestMode)
	userStore := mem.NewMemUserStore()
	planStore := mem.NewMemTrainingPlanStore()
	authSvc := service.NewAuthService(userStore)
	planSvc := service.NewTrainingPlanService(planStore)

	r := gin.New()
	storeCookie := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("rp.sid", storeCookie))

	api := r.Group("/api")
	RegisterAuthRoutes(api, authSvc)
	RegisterTrainingPlanRoutes(api, planSvc)

	return r, authSvc, planSvc
}

func loginAndGetCookies(t *testing.T, r *gin.Engine) []*http.Cookie {
	body := map[string]string{"email": "plans@example.com", "password": "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	return w.Result().Cookies()
}

func TestTrainingPlanController_Create(t *testing.T) {
	r, authSvc, _ := setupPlansTestRouter(t)
	_, err := authSvc.Register("plans@example.com", "password123")
	require.NoError(t, err)
	cookies := loginAndGetCookies(t, r)

	t.Run("creates plan when authenticated", func(t *testing.T) {
		body := map[string]interface{}{"name": "Marathon 2025", "endDate": "2025-06-15", "weeks": 16}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/", bytes.NewReader(bodyBytes))
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
		assert.Equal(t, "Marathon 2025", plan["name"])
		assert.Contains(t, plan["endDate"], "2025-06-15")
		assert.Equal(t, float64(16), plan["weeks"])
		assert.NotEmpty(t, plan["startDate"])
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		body := map[string]interface{}{"name": "Plan", "endDate": "2025-06-15", "weeks": 8}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid endDate returns 400", func(t *testing.T) {
		body := map[string]interface{}{"name": "Plan", "endDate": "invalid", "weeks": 8}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTrainingPlanController_GetByID(t *testing.T) {
	r, authSvc, planSvc := setupPlansTestRouter(t)
	u, _ := authSvc.Register("get@example.com", "password123")
	plan, _ := planSvc.Create(u.ID, "My Plan", mustParseDate("2025-05-01"), 8)

	body := map[string]string{"email": "get@example.com", "password": "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	t.Run("returns plan when owned by user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/plans/"+string(plan.ID), nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		p, ok := resp["plan"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "My Plan", p["name"])
	})

	t.Run("returns 404 for unknown plan", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/plans/nonexistent-id", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestTrainingPlanController_GetByUserID(t *testing.T) {
	r, authSvc, planSvc := setupPlansTestRouter(t)
	u, _ := authSvc.Register("get@example.com", "password123")
	_, _ = planSvc.Create(u.ID, "My Plan1", mustParseDate("2025-05-01"), 8)
	_, _ = planSvc.Create(u.ID, "My Plan2", mustParseDate("2025-05-01"), 8)

	body := map[string]string{"email": "get@example.com", "password": "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	t.Run("returns plans for authenticated user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/plans/", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		plans, ok := resp["plans"].([]interface{})
		require.True(t, ok)
		assert.Len(t, plans, 2)
		names := make([]string, len(plans))
		for i, p := range plans {
			names[i] = p.(map[string]interface{})["name"].(string)
		}
		assert.Contains(t, names, "My Plan1")
		assert.Contains(t, names, "My Plan2")
	})
}
