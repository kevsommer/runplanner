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

func setupPlansTestRouter(t *testing.T) (*gin.Engine, *service.AuthService, *service.TrainingPlanService, *service.WorkoutService) {
	gin.SetMode(gin.TestMode)
	userStore := mem.NewMemUserStore()
	planStore := mem.NewMemTrainingPlanStore()
	workoutStore := mem.NewMemWorkoutStore()
	authSvc := service.NewAuthService(userStore)
	planSvc := service.NewTrainingPlanService(planStore)
	workoutSvc := service.NewWorkoutService(workoutStore)

	r := gin.New()
	storeCookie := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("rp.sid", storeCookie))

	api := r.Group("/api")
	RegisterAuthRoutes(api, authSvc)
	RegisterTrainingPlanRoutes(api, planSvc, workoutSvc, nil)

	return r, authSvc, planSvc, workoutSvc
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
	r, authSvc, _, _ := setupPlansTestRouter(t)
	_, err := authSvc.Register("plans@example.com", "password123")
	require.NoError(t, err)
	cookies := loginAndGetCookies(t, r)

	t.Run("creates plan when authenticated", func(t *testing.T) {
		body := map[string]interface{}{"name": "Marathon 2025", "endDate": "2025-06-15", "weeks": 16}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans",bytes.NewReader(bodyBytes))
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
		req := httptest.NewRequest(http.MethodPost, "/api/plans",bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid endDate returns 400", func(t *testing.T) {
		body := map[string]interface{}{"name": "Plan", "endDate": "invalid", "weeks": 8}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans",bytes.NewReader(bodyBytes))
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
	r, authSvc, planSvc, _ := setupPlansTestRouter(t)
	u, _ := authSvc.Register("get@example.com", "password123")
	plan, _ := planSvc.Create(u.ID, "My Plan", mustParseDate("2025-05-01"), 8)

	body := map[string]string{"email": "get@example.com", "password": "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	t.Run("returns plan with weeksSummary", func(t *testing.T) {
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

		weeksSummary, ok := p["weeksSummary"].([]interface{})
		require.True(t, ok)
		assert.Len(t, weeksSummary, 8)

		week1 := weeksSummary[0].(map[string]interface{})
		assert.Equal(t, float64(1), week1["number"])
		days, ok := week1["days"].([]interface{})
		require.True(t, ok)
		assert.Len(t, days, 7)
		day1 := days[0].(map[string]interface{})
		assert.Equal(t, "Monday", day1["dayName"])
		workouts, ok := day1["workouts"].([]interface{})
		require.True(t, ok)
		assert.Len(t, workouts, 0)
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

func TestTrainingPlanController_GetByID_WithWorkouts(t *testing.T) {
	r, authSvc, planSvc, workoutSvc := setupPlansTestRouter(t)
	u, _ := authSvc.Register("detail@example.com", "password123")
	plan, _ := planSvc.Create(u.ID, "Test Plan", mustParseDate("2025-05-01"), 2)

	// Create workouts on week 1 Monday and Tuesday
	_, _ = workoutSvc.Create(plan.ID, "easy_run", plan.StartDate, "Monday run", 5.0)
	_, _ = workoutSvc.Create(plan.ID, "long_run", plan.StartDate.AddDate(0, 0, 1), "Tuesday run", 10.0)
	// Create a completed workout on week 1 Wednesday
	w3, _ := workoutSvc.Create(plan.ID, "tempo_run", plan.StartDate.AddDate(0, 0, 2), "Wednesday run", 8.0)
	w3.Status = "completed"
	_ = workoutSvc.Update(w3)

	body := map[string]string{"email": "detail@example.com", "password": "password123"}
	bodyBytes, _ := json.Marshal(body)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	lw := httptest.NewRecorder()
	r.ServeHTTP(lw, loginReq)
	cookies := lw.Result().Cookies()

	t.Run("weeksSummary includes workouts with correct km totals", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/plans/"+string(plan.ID), nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		p := resp["plan"].(map[string]interface{})
		weeksSummary := p["weeksSummary"].([]interface{})
		assert.Len(t, weeksSummary, 2)

		week1 := weeksSummary[0].(map[string]interface{})
		assert.Equal(t, float64(23), week1["plannedKm"])
		assert.Equal(t, float64(8), week1["doneKm"])
		assert.Equal(t, false, week1["allDone"])

		// Week 1 Monday should have 1 workout
		days := week1["days"].([]interface{})
		monday := days[0].(map[string]interface{})
		assert.Equal(t, "Monday", monday["dayName"])
		mondayWorkouts := monday["workouts"].([]interface{})
		assert.Len(t, mondayWorkouts, 1)

		// Week 2 should have no workouts
		week2 := weeksSummary[1].(map[string]interface{})
		assert.Equal(t, float64(0), week2["plannedKm"])
		assert.Equal(t, float64(0), week2["doneKm"])
		assert.Equal(t, false, week2["allDone"])
	})
}

func TestTrainingPlanController_Update(t *testing.T) {
	r, authSvc, planSvc, _ := setupPlansTestRouter(t)
	u, _ := authSvc.Register("update@example.com", "password123")
	plan, _ := planSvc.Create(u.ID, "Original Plan", mustParseDate("2025-06-15"), 8)

	body := map[string]string{"email": "update@example.com", "password": "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	t.Run("updates plan fields", func(t *testing.T) {
		body := map[string]interface{}{"name": "Updated Plan", "endDate": "2025-07-20", "weeks": 12}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/plans/"+string(plan.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		p := resp["plan"].(map[string]interface{})
		assert.Equal(t, "Updated Plan", p["name"])
		assert.Contains(t, p["endDate"], "2025-07-20")
		assert.Equal(t, float64(12), p["weeks"])
	})

	t.Run("returns 404 for unknown plan", func(t *testing.T) {
		body := map[string]interface{}{"name": "X", "endDate": "2025-07-20", "weeks": 8}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/plans/nonexistent", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 404 for other user's plan", func(t *testing.T) {
		other, _ := authSvc.Register("other-update@example.com", "password123")
		otherPlan, _ := planSvc.Create(other.ID, "Other Plan", mustParseDate("2025-06-15"), 8)

		body := map[string]interface{}{"name": "Hacked", "endDate": "2025-07-20", "weeks": 8}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/plans/"+string(otherPlan.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 400 for invalid endDate", func(t *testing.T) {
		body := map[string]interface{}{"name": "Plan", "endDate": "not-a-date", "weeks": 8}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/plans/"+string(plan.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 for missing fields", func(t *testing.T) {
		body := map[string]interface{}{"name": "Plan"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/plans/"+string(plan.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		body := map[string]interface{}{"name": "Plan", "endDate": "2025-07-20", "weeks": 8}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/plans/"+string(plan.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestTrainingPlanController_Delete(t *testing.T) {
	r, authSvc, planSvc, _ := setupPlansTestRouter(t)
	u, _ := authSvc.Register("delete@example.com", "password123")
	plan, _ := planSvc.Create(u.ID, "To Delete", mustParseDate("2025-06-15"), 8)

	body := map[string]string{"email": "delete@example.com", "password": "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	t.Run("deletes plan", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/plans/"+string(plan.ID), nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, true, resp["deleted"])

		// Verify plan is gone
		req = httptest.NewRequest(http.MethodGet, "/api/plans/"+string(plan.ID), nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 404 for unknown plan", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/plans/nonexistent", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 404 for other user's plan", func(t *testing.T) {
		other, _ := authSvc.Register("other-delete@example.com", "password123")
		otherPlan, _ := planSvc.Create(other.ID, "Other Plan", mustParseDate("2025-06-15"), 8)

		req := httptest.NewRequest(http.MethodDelete, "/api/plans/"+string(otherPlan.ID), nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/plans/some-id", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestTrainingPlanController_GetByUserID(t *testing.T) {
	r, authSvc, planSvc, _ := setupPlansTestRouter(t)
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
		req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
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
