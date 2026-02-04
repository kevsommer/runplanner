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
	planID := model.TrainingPlanID("plan-1")

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
	created1, err := svc.Create(planID, "easy_run", day, "5km easy run", 5.0)
	created2, err := svc.Create(planID, "tempo_run", day, "6km tempo run", 6.0)
	require.NoError(t, err)

	t.Run("returns plan by id", func(t *testing.T) {
		plan, err := svc.GetByPlanID(planID)
		require.NoError(t, err)
		assert.Len(t, plan, 2)
		assert.Equal(t, created1.ID, plan[0].ID)
		assert.Equal(t, created2.ID, plan[1].ID)
	})
}
