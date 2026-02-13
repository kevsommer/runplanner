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

func setupWorkoutsTestRouter(t *testing.T) (*gin.Engine, *service.AuthService, *service.TrainingPlanService, *service.WorkoutService) {
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
	RegisterTrainingPlanRoutes(api, planSvc)
	RegisterWorkoutRoutes(api, workoutSvc, planSvc)

	return r, authSvc, planSvc, workoutSvc
}

func loginAndGetWorkoutCookies(t *testing.T, r *gin.Engine, email, password string) []*http.Cookie {
	body := map[string]string{"email": email, "password": password}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	return w.Result().Cookies()
}

func TestWorkoutController_Create(t *testing.T) {
	r, authSvc, planSvc, _ := setupWorkoutsTestRouter(t)
	u, err := authSvc.Register("workouts@example.com", "password123")
	require.NoError(t, err)
	plan, err := planSvc.Create(u.ID, "My Plan", mustParseDate("2025-06-15"), 16)
	require.NoError(t, err)
	cookies := loginAndGetWorkoutCookies(t, r, "workouts@example.com", "password123")

	t.Run("creates workout when authenticated", func(t *testing.T) {
		body := map[string]interface{}{
			"planId":      string(plan.ID),
			"runType":     "easy_run",
			"day":         "2025-06-01",
			"description": "5km easy run",
			"distance":    5.0,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/workouts/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		workout, ok := resp["workout"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, string(plan.ID), workout["planId"])
		assert.Equal(t, "easy_run", workout["runType"])
		assert.Equal(t, 5.0, workout["distance"])
		assert.Equal(t, "5km easy run", workout["description"])
	})

	t.Run("creates workout with empty description when authenticated", func(t *testing.T) {
		body := map[string]interface{}{
			"planId":      string(plan.ID),
			"runType":     "easy_run",
			"day":         "2025-06-01",
			"description": "",
			"distance":    5.0,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/workouts/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		workout, ok := resp["workout"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, string(plan.ID), workout["planId"])
		assert.Equal(t, "easy_run", workout["runType"])
		assert.Equal(t, 5.0, workout["distance"])
		assert.Equal(t, "", workout["description"])
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		body := map[string]interface{}{
			"planId":      string(plan.ID),
			"runType":     "easy_run",
			"day":         "2025-06-01",
			"description": "5km easy run",
			"distance":    5.0,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/workouts/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid day returns 400", func(t *testing.T) {
		body := map[string]interface{}{
			"planId":      string(plan.ID),
			"runType":     "easy_run",
			"day":         "invalid",
			"description": "5km easy run",
			"distance":    5.0,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/workouts/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestWorkoutController_BulkCreate(t *testing.T) {
	r, authSvc, planSvc, _ := setupWorkoutsTestRouter(t)
	u, err := authSvc.Register("bulk@example.com", "password123")
	require.NoError(t, err)
	plan, err := planSvc.Create(u.ID, "Bulk Plan", mustParseDate("2025-06-15"), 16)
	require.NoError(t, err)
	cookies := loginAndGetWorkoutCookies(t, r, "bulk@example.com", "password123")

	t.Run("creates multiple workouts", func(t *testing.T) {
		body := map[string]interface{}{
			"workouts": []map[string]interface{}{
				{"runType": "easy_run", "week": 1, "dayOfWeek": 1, "description": "5km easy", "distance": 5.0},
				{"runType": "long_run", "week": 1, "dayOfWeek": 7, "description": "18km long", "distance": 18.0},
				{"runType": "tempo_run", "week": 2, "dayOfWeek": 3, "description": "", "distance": 7.5},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/"+string(plan.ID)+"/workouts/bulk", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		workouts, ok := resp["workouts"].([]interface{})
		require.True(t, ok)
		assert.Len(t, workouts, 3)

		first := workouts[0].(map[string]interface{})
		assert.Equal(t, "easy_run", first["runType"])
		assert.Equal(t, 5.0, first["distance"])
		assert.NotEmpty(t, first["id"])
		assert.Equal(t, string(plan.ID), first["planId"])
	})

	t.Run("workouts are retrievable after bulk create", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/plans/"+string(plan.ID)+"/workouts", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		workouts, ok := resp["workouts"].([]interface{})
		require.True(t, ok)
		assert.Len(t, workouts, 3)
	})

	t.Run("invalid run type returns 400 and creates nothing", func(t *testing.T) {
		otherPlan, _ := planSvc.Create(u.ID, "Other Plan", mustParseDate("2025-07-01"), 8)
		body := map[string]interface{}{
			"workouts": []map[string]interface{}{
				{"runType": "easy_run", "week": 1, "dayOfWeek": 1, "description": "ok", "distance": 5.0},
				{"runType": "sprint", "week": 1, "dayOfWeek": 2, "description": "bad", "distance": 3.0},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/"+string(otherPlan.ID)+"/workouts/bulk", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["error"], "workout[1]")

		// verify no workouts were created
		getReq := httptest.NewRequest(http.MethodGet, "/api/plans/"+string(otherPlan.ID)+"/workouts", nil)
		for _, c := range cookies {
			getReq.AddCookie(c)
		}
		getW := httptest.NewRecorder()
		r.ServeHTTP(getW, getReq)
		var getResp map[string]interface{}
		require.NoError(t, json.Unmarshal(getW.Body.Bytes(), &getResp))
		workouts := getResp["workouts"]
		assert.Nil(t, workouts)
	})

	t.Run("week exceeding plan weeks returns 400", func(t *testing.T) {
		shortPlan, _ := planSvc.Create(u.ID, "Short Plan", mustParseDate("2025-08-01"), 4)
		body := map[string]interface{}{
			"workouts": []map[string]interface{}{
				{"runType": "easy_run", "week": 5, "dayOfWeek": 1, "description": "too far", "distance": 5.0},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/"+string(shortPlan.ID)+"/workouts/bulk", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["error"], "week must be between")
	})

	t.Run("dayOfWeek out of range returns 400", func(t *testing.T) {
		body := map[string]interface{}{
			"workouts": []map[string]interface{}{
				{"runType": "easy_run", "week": 1, "dayOfWeek": 8, "description": "bad", "distance": 5.0},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/"+string(plan.ID)+"/workouts/bulk", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty workouts array returns 400", func(t *testing.T) {
		body := map[string]interface{}{}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/"+string(plan.ID)+"/workouts/bulk", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("nonexistent plan returns 404", func(t *testing.T) {
		body := map[string]interface{}{
			"workouts": []map[string]interface{}{
				{"runType": "easy_run", "week": 1, "dayOfWeek": 1, "description": "ok", "distance": 5.0},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/nonexistent/workouts/bulk", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("plan owned by another user returns 404", func(t *testing.T) {
		otherUser, _ := authSvc.Register("otherbulk@example.com", "password123")
		otherPlan, _ := planSvc.Create(otherUser.ID, "Secret Plan", mustParseDate("2025-06-01"), 8)

		body := map[string]interface{}{
			"workouts": []map[string]interface{}{
				{"runType": "easy_run", "week": 1, "dayOfWeek": 1, "description": "ok", "distance": 5.0},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/"+string(otherPlan.ID)+"/workouts/bulk", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		body := map[string]interface{}{
			"workouts": []map[string]interface{}{
				{"runType": "easy_run", "week": 1, "dayOfWeek": 1, "description": "ok", "distance": 5.0},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/plans/"+string(plan.ID)+"/workouts/bulk", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestWorkoutController_GetByID(t *testing.T) {
	r, authSvc, planSvc, workoutSvc := setupWorkoutsTestRouter(t)
	u, _ := authSvc.Register("getworkout@example.com", "password123")
	plan, _ := planSvc.Create(u.ID, "My Plan", mustParseDate("2025-05-01"), 8)
	created, _ := workoutSvc.Create(plan.ID, "easy_run", mustParseDate("2025-04-01"), "5km easy run", 5.0)

	cookies := loginAndGetWorkoutCookies(t, r, "getworkout@example.com", "password123")

	t.Run("returns workout when owned by user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workouts/"+string(created.ID), nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		workout, ok := resp["workout"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "5km easy run", workout["description"])
	})

	t.Run("returns 404 for unknown workout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workouts/nonexistent-id", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestWorkoutController_GetByPlanID(t *testing.T) {
	r, authSvc, planSvc, workoutSvc := setupWorkoutsTestRouter(t)
	u, _ := authSvc.Register("listworkouts@example.com", "password123")
	plan, _ := planSvc.Create(u.ID, "My Plan", mustParseDate("2025-05-01"), 8)

	_, _ = workoutSvc.Create(plan.ID, "easy_run", mustParseDate("2025-04-01"), "5km easy run", 5.0)
	_, _ = workoutSvc.Create(plan.ID, "tempo_run", mustParseDate("2025-04-02"), "6km tempo run", 6.0)

	cookies := loginAndGetWorkoutCookies(t, r, "listworkouts@example.com", "password123")

	t.Run("returns workouts for plan", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/plans/"+string(plan.ID)+"/workouts", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		workouts, ok := resp["workouts"].([]interface{})
		require.True(t, ok)
		assert.Len(t, workouts, 2)
	})
}

func TestWorkoutController_Update(t *testing.T) {
	r, authSvc, planSvc, workoutSvc := setupWorkoutsTestRouter(t)
	u, _ := authSvc.Register("updateworkout@example.com", "password123")
	plan, _ := planSvc.Create(u.ID, "My Plan", mustParseDate("2025-05-01"), 8)
	created, _ := workoutSvc.Create(plan.ID, "easy_run", mustParseDate("2025-04-01"), "5km easy run", 5.0)

	cookies := loginAndGetWorkoutCookies(t, r, "updateworkout@example.com", "password123")

	t.Run("updates workout fields", func(t *testing.T) {
		body := map[string]interface{}{
			"description": "Updated description",
			"notes":       "Great run",
			"status":      "completed",
			"distance":    10.0,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/workouts/"+string(created.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		workout, ok := resp["workout"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Updated description", workout["description"])
		assert.Equal(t, "Great run", workout["notes"])
		assert.Equal(t, "completed", workout["status"])
		assert.Equal(t, 10.0, workout["distance"])
	})

	t.Run("partial update only changes provided fields", func(t *testing.T) {
		body := map[string]interface{}{
			"notes": "New notes only",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/workouts/"+string(created.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		workout, ok := resp["workout"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "New notes only", workout["notes"])
		assert.Equal(t, "Updated description", workout["description"]) // unchanged from previous test
	})

	t.Run("returns 404 for unknown workout", func(t *testing.T) {
		body := map[string]interface{}{"notes": "test"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/workouts/nonexistent-id", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 404 for workout owned by another user", func(t *testing.T) {
		otherUser, _ := authSvc.Register("other@example.com", "password123")
		otherPlan, _ := planSvc.Create(otherUser.ID, "Other Plan", mustParseDate("2025-05-01"), 8)
		otherWorkout, _ := workoutSvc.Create(otherPlan.ID, "easy_run", mustParseDate("2025-04-01"), "other workout", 3.0)

		body := map[string]interface{}{"notes": "hacked"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/workouts/"+string(otherWorkout.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid run type returns 400", func(t *testing.T) {
		body := map[string]interface{}{"runType": "sprint"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/workouts/"+string(created.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("negative distance returns 400", func(t *testing.T) {
		body := map[string]interface{}{"distance": -1.0}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/workouts/"+string(created.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		body := map[string]interface{}{"notes": "test"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/workouts/"+string(created.ID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestWorkoutController_Delete(t *testing.T) {
	r, authSvc, planSvc, workoutSvc := setupWorkoutsTestRouter(t)
	u, _ := authSvc.Register("deleteworkout@example.com", "password123")
	plan, _ := planSvc.Create(u.ID, "My Plan", mustParseDate("2025-05-01"), 8)

	cookies := loginAndGetWorkoutCookies(t, r, "deleteworkout@example.com", "password123")

	t.Run("deletes workout when owned by user", func(t *testing.T) {
		created, _ := workoutSvc.Create(plan.ID, "easy_run", mustParseDate("2025-04-01"), "5km easy run", 5.0)

		req := httptest.NewRequest(http.MethodDelete, "/api/workouts/"+string(created.ID), nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, true, resp["deleted"])

		// Verify it's gone
		getReq := httptest.NewRequest(http.MethodGet, "/api/workouts/"+string(created.ID), nil)
		for _, c := range cookies {
			getReq.AddCookie(c)
		}
		getW := httptest.NewRecorder()
		r.ServeHTTP(getW, getReq)
		assert.Equal(t, http.StatusNotFound, getW.Code)
	})

	t.Run("returns 404 for unknown workout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/workouts/nonexistent-id", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 404 for workout owned by another user", func(t *testing.T) {
		otherUser, _ := authSvc.Register("otherdelete@example.com", "password123")
		otherPlan, _ := planSvc.Create(otherUser.ID, "Other Plan", mustParseDate("2025-05-01"), 8)
		otherWorkout, _ := workoutSvc.Create(otherPlan.ID, "easy_run", mustParseDate("2025-04-01"), "other workout", 3.0)

		req := httptest.NewRequest(http.MethodDelete, "/api/workouts/"+string(otherWorkout.ID), nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		created, _ := workoutSvc.Create(plan.ID, "easy_run", mustParseDate("2025-04-01"), "to delete", 5.0)

		req := httptest.NewRequest(http.MethodDelete, "/api/workouts/"+string(created.ID), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

