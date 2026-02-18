package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kevsommer/runplanner/internal/ai"
	"github.com/kevsommer/runplanner/internal/model"
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

func validAIResponse() string {
	return `{"workouts": [
		{"runType": "easy_run", "week": 1, "dayOfWeek": 1, "description": "Easy 5k", "distance": 5.0},
		{"runType": "easy_run", "week": 1, "dayOfWeek": 3, "description": "Easy 6k", "distance": 6.0},
		{"runType": "long_run", "week": 1, "dayOfWeek": 6, "description": "Long run", "distance": 12.0}
	]}`
}

func validInput() GenerateInput {
	return GenerateInput{
		Name:          "Marathon Plan",
		EndDate:       time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
		Weeks:         8,
		BaseKmPerWeek: 30,
		RunsPerWeek:   3,
	}
}

func setupGenerateTest(mockClient ai.Client) (*GenerateService, *TrainingPlanService, *WorkoutService) {
	planStore := mem.NewMemTrainingPlanStore()
	workoutStore := mem.NewMemWorkoutStore()
	planSvc := NewTrainingPlanService(planStore)
	workoutSvc := NewWorkoutService(workoutStore)
	genSvc := NewGenerateService(mockClient, planSvc, workoutSvc)
	return genSvc, planSvc, workoutSvc
}

func TestGenerateService_Generate(t *testing.T) {
	t.Run("creates plan and workouts from AI response", func(t *testing.T) {
		mock := &mockAIClient{response: validAIResponse()}
		genSvc, _, workoutSvc := setupGenerateTest(mock)

		plan, workouts, err := genSvc.Generate(context.Background(), model.UserID("user1"), validInput())
		require.NoError(t, err)
		assert.Equal(t, "Marathon Plan", plan.Name)
		assert.Equal(t, 8, plan.Weeks)
		assert.Len(t, workouts, 3)
		assert.Equal(t, "easy_run", workouts[0].RunType)
		assert.Equal(t, 5.0, workouts[0].Distance)
		assert.Equal(t, "long_run", workouts[2].RunType)

		// Verify workouts are persisted
		stored, err := workoutSvc.GetByPlanID(plan.ID)
		require.NoError(t, err)
		assert.Len(t, stored, 3)
	})

	t.Run("returns error when AI client is nil", func(t *testing.T) {
		genSvc, _, _ := setupGenerateTest(nil)

		_, _, err := genSvc.Generate(context.Background(), model.UserID("user1"), validInput())
		assert.ErrorIs(t, err, ErrAINotConfigured)
	})

	t.Run("returns error when AI client fails", func(t *testing.T) {
		mock := &mockAIClient{err: errors.New("connection timeout")}
		genSvc, _, _ := setupGenerateTest(mock)

		_, _, err := genSvc.Generate(context.Background(), model.UserID("user1"), validInput())
		assert.ErrorIs(t, err, ErrAIGeneration)
		assert.Contains(t, err.Error(), "connection timeout")
	})

	t.Run("returns error when AI response is invalid JSON", func(t *testing.T) {
		mock := &mockAIClient{response: "not json at all"}
		genSvc, _, _ := setupGenerateTest(mock)

		_, _, err := genSvc.Generate(context.Background(), model.UserID("user1"), validInput())
		assert.ErrorIs(t, err, ErrAIGeneration)
		assert.Contains(t, err.Error(), "parse AI response")
	})

	t.Run("returns error when AI response has empty workouts", func(t *testing.T) {
		mock := &mockAIClient{response: `{"workouts": []}`}
		genSvc, _, _ := setupGenerateTest(mock)

		_, _, err := genSvc.Generate(context.Background(), model.UserID("user1"), validInput())
		assert.ErrorIs(t, err, ErrAIGeneration)
		assert.Contains(t, err.Error(), "no workouts")
	})

	t.Run("cleans up plan when workout creation fails", func(t *testing.T) {
		// AI returns a workout with invalid run type — CreateBatch will reject it
		mock := &mockAIClient{response: `{"workouts": [
			{"runType": "invalid_type", "week": 1, "dayOfWeek": 1, "description": "bad", "distance": 5.0}
		]}`}
		genSvc, planSvc, _ := setupGenerateTest(mock)

		_, _, err := genSvc.Generate(context.Background(), model.UserID("user1"), validInput())
		require.Error(t, err)

		// Verify no plans remain for this user
		plans, err := planSvc.GetByUserID(model.UserID("user1"))
		require.NoError(t, err)
		assert.Len(t, plans, 0)
	})
}

func TestValidateGenerateInput(t *testing.T) {
	t.Run("valid input passes", func(t *testing.T) {
		err := validateGenerateInput(validInput())
		assert.NoError(t, err)
	})

	t.Run("weeks less than 6 fails", func(t *testing.T) {
		input := validInput()
		input.Weeks = 5
		err := validateGenerateInput(input)
		assert.ErrorIs(t, err, ErrInvalidInput)
		assert.Contains(t, err.Error(), "weeks must be at least 6")
	})

	t.Run("zero baseKmPerWeek fails", func(t *testing.T) {
		input := validInput()
		input.BaseKmPerWeek = 0
		err := validateGenerateInput(input)
		assert.ErrorIs(t, err, ErrInvalidInput)
		assert.Contains(t, err.Error(), "baseKmPerWeek")
	})

	t.Run("negative baseKmPerWeek fails", func(t *testing.T) {
		input := validInput()
		input.BaseKmPerWeek = -10
		err := validateGenerateInput(input)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("runsPerWeek below 2 fails", func(t *testing.T) {
		input := validInput()
		input.RunsPerWeek = 1
		err := validateGenerateInput(input)
		assert.ErrorIs(t, err, ErrInvalidInput)
		assert.Contains(t, err.Error(), "runsPerWeek")
	})

	t.Run("runsPerWeek above 7 fails", func(t *testing.T) {
		input := validInput()
		input.RunsPerWeek = 8
		err := validateGenerateInput(input)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("empty name fails", func(t *testing.T) {
		input := validInput()
		input.Name = ""
		err := validateGenerateInput(input)
		assert.ErrorIs(t, err, ErrInvalidInput)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("boundary: 6 weeks and 2 runs/week passes", func(t *testing.T) {
		input := validInput()
		input.Weeks = 6
		input.RunsPerWeek = 2
		err := validateGenerateInput(input)
		assert.NoError(t, err)
	})

	t.Run("boundary: 7 runs/week passes", func(t *testing.T) {
		input := validInput()
		input.RunsPerWeek = 7
		err := validateGenerateInput(input)
		assert.NoError(t, err)
	})
}

func TestParseWorkouts(t *testing.T) {
	t.Run("parses valid response", func(t *testing.T) {
		items, err := parseWorkouts(validAIResponse())
		require.NoError(t, err)
		assert.Len(t, items, 3)
		assert.Equal(t, "easy_run", items[0].RunType)
		assert.Equal(t, 1, items[0].Week)
		assert.Equal(t, 1, items[0].DayOfWeek)
		assert.Equal(t, "Easy 5k", items[0].Description)
		assert.Equal(t, 5.0, items[0].Distance)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		_, err := parseWorkouts("{bad json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")
	})

	t.Run("rejects empty workouts array", func(t *testing.T) {
		_, err := parseWorkouts(`{"workouts": []}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no workouts")
	})

	t.Run("rejects missing workouts key", func(t *testing.T) {
		_, err := parseWorkouts(`{"plans": []}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no workouts")
	})

	t.Run("rounds distances to whole numbers", func(t *testing.T) {
		raw := `{"workouts": [
			{"runType": "easy_run", "week": 1, "dayOfWeek": 1, "description": "easy", "distance": 9.6},
			{"runType": "long_run", "week": 1, "dayOfWeek": 6, "description": "long", "distance": 18.3}
		]}`
		items, err := parseWorkouts(raw)
		require.NoError(t, err)
		assert.Equal(t, 10.0, items[0].Distance)
		assert.Equal(t, 18.0, items[1].Distance)
	})
}
