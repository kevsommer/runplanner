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

func TestBuildPlanDetail(t *testing.T) {
	plan := &model.TrainingPlan{
		ID:        "plan-1",
		UserID:    "user-1",
		Name:      "Test Plan",
		StartDate: time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC), // Monday
		EndDate:   time.Date(2025, 3, 23, 0, 0, 0, 0, time.UTC),
		Weeks:     2,
	}

	t.Run("empty workouts produces correct week structure", func(t *testing.T) {
		detail := BuildPlanDetail(plan, []*model.Workout{})

		assert.Equal(t, plan.ID, detail.ID)
		assert.Equal(t, plan.Name, detail.Name)
		require.Len(t, detail.WeeksSummary, 2)

		week1 := detail.WeeksSummary[0]
		assert.Equal(t, 1, week1.Number)
		assert.Equal(t, 0.0, week1.PlannedKm)
		assert.Equal(t, 0.0, week1.DoneKm)
		assert.False(t, week1.AllDone)
		require.Len(t, week1.Days, 7)
		assert.Equal(t, "Monday", week1.Days[0].DayName)
		assert.Equal(t, "2025-03-10", week1.Days[0].Date)
		assert.Equal(t, "Sunday", week1.Days[6].DayName)
		assert.Equal(t, "2025-03-16", week1.Days[6].Date)
		assert.Len(t, week1.Days[0].Workouts, 0)

		week2 := detail.WeeksSummary[1]
		assert.Equal(t, 2, week2.Number)
		assert.Equal(t, "2025-03-17", week2.Days[0].Date)
	})

	t.Run("workouts are assigned to correct days", func(t *testing.T) {
		workouts := []*model.Workout{
			{ID: "w1", PlanID: "plan-1", RunType: "easy_run", Day: time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC), Status: "pending", Distance: 5},
			{ID: "w2", PlanID: "plan-1", RunType: "long_run", Day: time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC), Status: "completed", Distance: 10},
			{ID: "w3", PlanID: "plan-1", RunType: "intervals", Day: time.Date(2025, 3, 17, 0, 0, 0, 0, time.UTC), Status: "completed", Distance: 7},
		}

		detail := BuildPlanDetail(plan, workouts)

		// Week 1: 2 workouts on Monday
		week1 := detail.WeeksSummary[0]
		assert.Len(t, week1.Days[0].Workouts, 2)
		assert.Equal(t, 15.0, week1.PlannedKm)
		assert.Equal(t, 10.0, week1.DoneKm)
		assert.False(t, week1.AllDone)

		// Week 2: 1 workout on Monday
		week2 := detail.WeeksSummary[1]
		assert.Len(t, week2.Days[0].Workouts, 1)
		assert.Equal(t, 7.0, week2.PlannedKm)
		assert.Equal(t, 7.0, week2.DoneKm)
		assert.True(t, week2.AllDone)
	})

	t.Run("allDone is false when no workouts exist", func(t *testing.T) {
		detail := BuildPlanDetail(plan, []*model.Workout{})
		assert.False(t, detail.WeeksSummary[0].AllDone)
	})

	t.Run("allDone is true only when all workouts completed", func(t *testing.T) {
		workouts := []*model.Workout{
			{ID: "w1", PlanID: "plan-1", RunType: "easy_run", Day: time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC), Status: "completed", Distance: 5},
			{ID: "w2", PlanID: "plan-1", RunType: "long_run", Day: time.Date(2025, 3, 11, 0, 0, 0, 0, time.UTC), Status: "completed", Distance: 10},
		}
		detail := BuildPlanDetail(plan, workouts)
		assert.True(t, detail.WeeksSummary[0].AllDone)
	})

	t.Run("skipped workouts count toward planned but not done km", func(t *testing.T) {
		workouts := []*model.Workout{
			{ID: "w1", PlanID: "plan-1", RunType: "easy_run", Day: time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC), Status: "skipped", Distance: 5},
		}
		detail := BuildPlanDetail(plan, workouts)
		week1 := detail.WeeksSummary[0]
		assert.Equal(t, 5.0, week1.PlannedKm)
		assert.Equal(t, 0.0, week1.DoneKm)
		assert.False(t, week1.AllDone)
	})
}

func TestTrainingPlanService_GetByUserID(t *testing.T) {
	svc := setupTrainingPlanTest(t)
	userID := model.UserID("user-2")
	created1, err := svc.Create(userID, "Test Plan 1", time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC), 8)
	created2, err := svc.Create(userID, "Test Plan 2", time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC), 8)
	require.NoError(t, err)

	t.Run("returns plan by id", func(t *testing.T) {
		plan, err := svc.GetByUserID(userID)
		require.NoError(t, err)
		assert.Len(t, plan, 2)
		assert.Equal(t, created1.ID, plan[0].ID)
		assert.Equal(t, created2.ID, plan[1].ID)
	})
}
