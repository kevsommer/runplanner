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

func setupTrainingPlanTest(t *testing.T) *TrainingPlanService {
	return NewTrainingPlanService(mem.NewMemTrainingPlanStore())
}

func TestStartDateFor(t *testing.T) {
	// Race on Saturday 2025-04-12, 12 weeks -> Monday of week 1
	endDate := time.Date(2025, 4, 12, 0, 0, 0, 0, time.UTC) // Saturday
	start := StartDateFor(endDate, 12)
	// Week 12 contains Apr 12. Monday of week 12 = Apr 7.
	// Monday of week 1 = Apr 7 - 11*7 = Apr 7 - 77 days = Jan 20
	assert.Equal(t, time.Monday, start.Weekday())
	assert.Equal(t, 2025, start.Year())
	assert.Equal(t, time.January, start.Month())
	assert.Equal(t, 20, start.Day())

	// Race on Monday 2025-05-05, 4 weeks -> Monday of week 1 = Apr 14 (3 weeks before May 5)
	endDate = time.Date(2025, 5, 5, 0, 0, 0, 0, time.UTC) // Monday
	start = StartDateFor(endDate, 4)
	assert.Equal(t, time.Monday, start.Weekday())
	assert.Equal(t, 2025, start.Year())
	assert.Equal(t, time.April, start.Month())
	assert.Equal(t, 14, start.Day())
}

func TestTrainingPlanService_Create(t *testing.T) {
	svc := setupTrainingPlanTest(t)
	userID := model.UserID("user-1")

	t.Run("creates plan with calculated start date", func(t *testing.T) {
		endDate := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC) // Sunday
		plan, err := svc.Create(userID, "Marathon 2025", endDate, 16)

		require.NoError(t, err)
		require.NotNil(t, plan)
		assert.NotEmpty(t, plan.ID)
		assert.Equal(t, userID, plan.UserID)
		assert.Equal(t, "Marathon 2025", plan.Name)
		assert.Equal(t, endDate, plan.EndDate)
		assert.Equal(t, 16, plan.Weeks)
		assert.Equal(t, time.Monday, plan.StartDate.Weekday())
		assert.True(t, plan.StartDate.Before(plan.EndDate))
	})

	t.Run("empty name returns ErrInvalidName", func(t *testing.T) {
		plan, err := svc.Create(userID, "", time.Now(), 8)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidName, err)
		assert.Nil(t, plan)
	})

	t.Run("weeks < 1 returns ErrInvalidWeeks", func(t *testing.T) {
		plan, err := svc.Create(userID, "Plan", time.Now(), 0)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidWeeks, err)
		assert.Nil(t, plan)
	})
}

func TestTrainingPlanService_GetByID(t *testing.T) {
	svc := setupTrainingPlanTest(t)
	userID := model.UserID("user-1")
	created, err := svc.Create(userID, "Test Plan", time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC), 8)
	require.NoError(t, err)

	t.Run("returns plan by id", func(t *testing.T) {
		plan, err := svc.GetByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, plan.ID)
		assert.Equal(t, "Test Plan", plan.Name)
	})

	t.Run("unknown id returns ErrNotFound", func(t *testing.T) {
		plan, err := svc.GetByID("nonexistent")
		assert.Error(t, err)
		assert.Equal(t, store.ErrNotFound, err)
		assert.Nil(t, plan)
	})
}

