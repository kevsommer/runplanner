package service

import (
	"testing"
	"time"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
	"github.com/kevsommer/runplanner/internal/store/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWorkoutTest(t *testing.T) *WorkoutService {
	return NewWorkoutService(mem.NewMemWorkoutStore())
}

func TestWorkoutService_Create(t *testing.T) {
	svc := setupWorkoutTest(t)
	planID := model.TrainingPlanID("random-plan-id")

	t.Run("creates workout", func(t *testing.T) {
		day := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
		workout, err := svc.Create(planID, "easy_run", day, "5km easy run", 5.0)

		require.NoError(t, err)
		require.NotNil(t, workout)
		assert.NotEmpty(t, workout.ID)
		assert.Equal(t, planID, workout.PlanID)
		assert.Equal(t, "easy_run", workout.RunType)
		assert.Equal(t, day, workout.Day)
		assert.Equal(t, 5.0, workout.Distance)
		assert.Equal(t, false, workout.Done)
		assert.Equal(t, "5km easy run", workout.Description)
		assert.Equal(t, "", workout.Notes)
	})

	t.Run("creates workout with empty description", func(t *testing.T) {
		day := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
		workout, err := svc.Create(planID, "easy_run", day, "", 5.0)

		require.NoError(t, err)
		require.NotNil(t, workout)
		assert.Equal(t, "", workout.Description)
	})

	t.Run("invalid run type returns ErrInvalidRunType", func(t *testing.T) {
		day := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

		workout, err := svc.Create(planID, "invalid_run_type", day, "5km easy run", 5.0)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidRunType, err)
		assert.Nil(t, workout)
	})

	t.Run("distance < 0 returns ErrInvalidDistance", func(t *testing.T) {
		day := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

		workout, err := svc.Create(planID, "easy_run", day, "5km easy run", -1.0)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidDistance, err)
		assert.Nil(t, workout)
	})
}

func TestWorkoutService_CreateBatch(t *testing.T) {
	svc := setupWorkoutTest(t)
	// Plan starts Monday 2025-03-10, 12 weeks
	plan := &model.TrainingPlan{
		ID:        "plan-batch",
		StartDate: time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC),
		Weeks:     12,
	}

	t.Run("creates all workouts with correct dates", func(t *testing.T) {
		items := []BulkWorkoutInput{
			{RunType: "easy_run", Week: 1, DayOfWeek: 1, Description: "5km easy", Distance: 5.0},       // Mon week 1 = 2025-03-10
			{RunType: "long_run", Week: 1, DayOfWeek: 7, Description: "18km long", Distance: 18.0},      // Sun week 1 = 2025-03-16
			{RunType: "tempo_run", Week: 2, DayOfWeek: 3, Description: "tempo", Distance: 8.0},           // Wed week 2 = 2025-03-19
		}
		workouts, err := svc.CreateBatch(plan, items)

		require.NoError(t, err)
		require.Len(t, workouts, 3)
		for i, w := range workouts {
			assert.NotEmpty(t, w.ID)
			assert.Equal(t, plan.ID, w.PlanID)
			assert.Equal(t, items[i].RunType, w.RunType)
			assert.Equal(t, items[i].Distance, w.Distance)
			assert.Equal(t, items[i].Description, w.Description)
			assert.False(t, w.Done)
		}
		assert.Equal(t, time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC), workouts[0].Day)
		assert.Equal(t, time.Date(2025, 3, 16, 0, 0, 0, 0, time.UTC), workouts[1].Day)
		assert.Equal(t, time.Date(2025, 3, 19, 0, 0, 0, 0, time.UTC), workouts[2].Day)

		// verify they're retrievable
		stored, err := svc.GetByPlanID(plan.ID)
		require.NoError(t, err)
		assert.Len(t, stored, 3)
	})

	t.Run("empty description is allowed", func(t *testing.T) {
		emptyPlan := &model.TrainingPlan{ID: "plan-empty-desc", StartDate: plan.StartDate, Weeks: 12}
		items := []BulkWorkoutInput{
			{RunType: "easy_run", Week: 1, DayOfWeek: 1, Description: "", Distance: 6.0},
		}
		workouts, err := svc.CreateBatch(emptyPlan, items)
		require.NoError(t, err)
		assert.Equal(t, "", workouts[0].Description)
	})

	t.Run("invalid run type returns BatchValidationError", func(t *testing.T) {
		items := []BulkWorkoutInput{
			{RunType: "easy_run", Week: 1, DayOfWeek: 1, Description: "ok", Distance: 5.0},
			{RunType: "sprint", Week: 1, DayOfWeek: 2, Description: "bad", Distance: 3.0},
		}
		workouts, err := svc.CreateBatch(plan, items)

		assert.Nil(t, workouts)
		require.Error(t, err)
		bve, ok := err.(*BatchValidationError)
		require.True(t, ok)
		assert.Equal(t, 1, bve.Index)
		assert.Contains(t, bve.Message, "invalid run type")
	})

	t.Run("negative distance returns BatchValidationError", func(t *testing.T) {
		items := []BulkWorkoutInput{
			{RunType: "easy_run", Week: 1, DayOfWeek: 1, Description: "bad", Distance: -1.0},
		}
		workouts, err := svc.CreateBatch(plan, items)

		assert.Nil(t, workouts)
		require.Error(t, err)
		bve, ok := err.(*BatchValidationError)
		require.True(t, ok)
		assert.Equal(t, 0, bve.Index)
		assert.Contains(t, bve.Message, "distance")
	})

	t.Run("week exceeding plan weeks returns BatchValidationError", func(t *testing.T) {
		items := []BulkWorkoutInput{
			{RunType: "easy_run", Week: 13, DayOfWeek: 1, Description: "too far", Distance: 5.0},
		}
		workouts, err := svc.CreateBatch(plan, items)

		assert.Nil(t, workouts)
		require.Error(t, err)
		bve, ok := err.(*BatchValidationError)
		require.True(t, ok)
		assert.Equal(t, 0, bve.Index)
		assert.Contains(t, bve.Message, "week must be between 1 and 12")
	})

	t.Run("week 0 returns BatchValidationError", func(t *testing.T) {
		items := []BulkWorkoutInput{
			{RunType: "easy_run", Week: 0, DayOfWeek: 1, Description: "bad", Distance: 5.0},
		}
		workouts, err := svc.CreateBatch(plan, items)

		assert.Nil(t, workouts)
		require.Error(t, err)
		bve, ok := err.(*BatchValidationError)
		require.True(t, ok)
		assert.Contains(t, bve.Message, "week must be between")
	})

	t.Run("dayOfWeek out of range returns BatchValidationError", func(t *testing.T) {
		items := []BulkWorkoutInput{
			{RunType: "easy_run", Week: 1, DayOfWeek: 8, Description: "bad", Distance: 5.0},
		}
		workouts, err := svc.CreateBatch(plan, items)

		assert.Nil(t, workouts)
		require.Error(t, err)
		bve, ok := err.(*BatchValidationError)
		require.True(t, ok)
		assert.Contains(t, bve.Message, "dayOfWeek must be between")
	})

	t.Run("no workouts created when validation fails", func(t *testing.T) {
		isolatedPlan := &model.TrainingPlan{ID: "plan-no-create", StartDate: plan.StartDate, Weeks: 12}
		items := []BulkWorkoutInput{
			{RunType: "easy_run", Week: 1, DayOfWeek: 1, Description: "good", Distance: 5.0},
			{RunType: "bad_type", Week: 1, DayOfWeek: 2, Description: "bad", Distance: 3.0},
		}
		_, err := svc.CreateBatch(isolatedPlan, items)
		require.Error(t, err)

		stored, err := svc.GetByPlanID(isolatedPlan.ID)
		require.NoError(t, err)
		assert.Len(t, stored, 0)
	})
}

func TestWorkoutService_GetByID(t *testing.T) {
	svc := setupWorkoutTest(t)
	planID := model.TrainingPlanID("plan-1")

	day := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	created, err := svc.Create(planID, "easy_run", day, "5km easy run", 5.0)
	require.NoError(t, err)

	t.Run("returns plan by id", func(t *testing.T) {
		workout, err := svc.GetByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, workout.ID)
		assert.Equal(t, created.RunType, workout.RunType)
	})

	t.Run("unknown id returns ErrNotFound", func(t *testing.T) {
		workout, err := svc.GetByID("nonexistent")
		assert.Error(t, err)
		assert.Equal(t, store.ErrNotFound, err)
		assert.Nil(t, workout)
	})
}

func TestWorkoutService_GetByPlanID(t *testing.T) {
	svc := setupWorkoutTest(t)
	planID := model.TrainingPlanID("plan-1")
	day := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	created1, err1 := svc.Create(planID, "easy_run", day, "5km easy run", 5.0)
	created2, err2 := svc.Create(planID, "tempo_run", day, "6km tempo run", 6.0)
	require.NoError(t, err1)
	require.NoError(t, err2)

	t.Run("returns plan by id", func(t *testing.T) {
		plan, err := svc.GetByPlanID(planID)
		require.NoError(t, err)
		assert.Len(t, plan, 2)
		assert.Equal(t, created1.ID, plan[0].ID)
		assert.Equal(t, created2.ID, plan[1].ID)
	})
}

func TestWorkoutService_Update(t *testing.T) {
	svc := setupWorkoutTest(t)
	planID := model.TrainingPlanID("plan-1")
	day := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	created, err := svc.Create(planID, "easy_run", day, "5km easy run", 5.0)
	require.NoError(t, err)

	t.Run("updates workout", func(t *testing.T) {
		created.Description = "Updated description"
		created.Notes = "Some notes"
		created.Done = true
		err := svc.Update(created)
		require.NoError(t, err)

		updated, err := svc.GetByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", updated.Description)
		assert.Equal(t, "Some notes", updated.Notes)
		assert.Equal(t, true, updated.Done)
	})

	t.Run("invalid run type returns ErrInvalidRunType", func(t *testing.T) {
		created.RunType = "invalid_run_type"
		err := svc.Update(created)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidRunType, err)
	})

	t.Run("distance < 0 returns ErrInvalidDistance", func(t *testing.T) {
		created.RunType = "easy_run"
		created.Distance = -1.0
		err := svc.Update(created)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidDistance, err)
	})

	t.Run("unknown id returns ErrNotFound", func(t *testing.T) {
		unknown := &model.Workout{
			ID:          "nonexistent",
			PlanID:      planID,
			RunType:     "easy_run",
			Day:         day,
			Description: "Nonexistent workout",
			Notes:       "",
			Done:        false,
			Distance:    5.0,
		}
		err := svc.Update(unknown)
		assert.Error(t, err)
		assert.Equal(t, store.ErrNotFound, err)
	})
}

func TestWorkoutService_Delete(t *testing.T) {
	svc := setupWorkoutTest(t)
	planID := model.TrainingPlanID("plan-1")
	day := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	created, err := svc.Create(planID, "easy_run", day, "5km easy run", 5.0)
	require.NoError(t, err)

	t.Run("deletes workout", func(t *testing.T) {
		err := svc.Delete(created.ID)
		require.NoError(t, err)

		workout, err := svc.GetByID(created.ID)
		assert.Error(t, err)
		assert.Equal(t, store.ErrNotFound, err)
		assert.Nil(t, workout)
	})

	t.Run("unknown id returns ErrNotFound", func(t *testing.T) {
		err := svc.Delete("nonexistent")
		assert.Error(t, err)
		assert.Equal(t, store.ErrNotFound, err)
	})
}
