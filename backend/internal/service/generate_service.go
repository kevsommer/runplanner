package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/kevsommer/runplanner/internal/ai"
	"github.com/kevsommer/runplanner/internal/model"
)

var (
	ErrAINotConfigured = errors.New("AI generation is not configured")
	ErrAIGeneration    = errors.New("AI generation failed")
	ErrInvalidInput    = errors.New("invalid input")
)

type GenerateInput struct {
	Name          string
	EndDate       time.Time
	Weeks         int
	BaseKmPerWeek float64
	RunsPerWeek   int
}

type GenerateService struct {
	ai       ai.Client
	plans    *TrainingPlanService
	workouts *WorkoutService
}

func NewGenerateService(aiClient ai.Client, plans *TrainingPlanService, workouts *WorkoutService) *GenerateService {
	return &GenerateService{
		ai:       aiClient,
		plans:    plans,
		workouts: workouts,
	}
}

func (s *GenerateService) Generate(ctx context.Context, userID model.UserID, input GenerateInput) (*model.TrainingPlan, []*model.Workout, error) {
	if s.ai == nil {
		return nil, nil, ErrAINotConfigured
	}

	if err := validateGenerateInput(input); err != nil {
		return nil, nil, err
	}

	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(input)

	raw, err := s.ai.Complete(ctx, ai.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrAIGeneration, err)
	}

	items, err := parseWorkouts(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to parse AI response: %v", ErrAIGeneration, err)
	}

	plan, err := s.plans.Create(userID, input.Name, input.EndDate, input.Weeks)
	if err != nil {
		return nil, nil, err
	}

	workouts, err := s.workouts.CreateBatch(plan, items)
	if err != nil {
		_ = s.plans.Delete(plan.ID)
		return nil, nil, fmt.Errorf("failed to create workouts: %w", err)
	}

	return plan, workouts, nil
}

func validateGenerateInput(input GenerateInput) error {
	if input.Weeks < 6 {
		return fmt.Errorf("%w: weeks must be at least 6", ErrInvalidInput)
	}
	if input.BaseKmPerWeek <= 0 {
		return fmt.Errorf("%w: baseKmPerWeek must be greater than 0", ErrInvalidInput)
	}
	if input.RunsPerWeek < 2 || input.RunsPerWeek > 7 {
		return fmt.Errorf("%w: runsPerWeek must be between 2 and 7", ErrInvalidInput)
	}
	if input.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	return nil
}

func buildSystemPrompt() string {
	return `You are a running coach that creates structured training plans. You output JSON only.
{ "workouts": [
    { "runType": "easy_run", "week": 1, "dayOfWeek": 1, "description": "+4x 20s Strides", "distance": 8.0 },
    { "runType": "tempo_run", "week": 1, "dayOfWeek": 3, "description": "4k Easy\n3k Tempo\n3k Easy", "distance": 10.0 },
    { "runType": "easy_run", "week": 1, "dayOfWeek": 4, "description": "", "distance": 6.0 },
    { "runType": "long_run", "week": 1, "dayOfWeek": 6, "description": "All Easy", "distance": 16.0 }
  ] 
}

This is the json structure you MUST follow. Do NOT include any text outside the JSON object. All fields are required.
Rules for generating training plans:
- Every 4th week is a DELOAD week: reduce total volume by 40%
- The last 3 weeks before race day are TAPER weeks: progressively reduce volume (75%, 50%, 30% of peak)
- Apply ~10% weekly volume progression on non-deload, non-taper weeks
- Every week MUST include exactly one long_run, scheduled on a weekend (Saturday=6 or Sunday=7)
- Long run should start at 15-18 km in week 1 and progressively increase, peaking at 32-35 km for higher base volumes (40+ km/week) or 28-32 km for lower base volumes
- The long run MUST surpass 21 km (half marathon) well before the final weeks
- If runsPerWeek >= 3, include one speed session per week (tempo_run or intervals, alternating)
- Remaining runs should be easy_run
- 80% of weekly volume should come from easy_run and long_run combined; speed sessions are shorter
- The long_run should be the longest run of the week by a clear margin (e.g. 30-40% of weekly volume)
- Vary distances between runs — do NOT give every easy_run the same distance; mix shorter and longer easy days
- All distances MUST be whole integers (e.g. 8, 12, 15), never decimals
- Each workout needs a brief description

Valid run types: easy_run, intervals, long_run, tempo_run, race
DayOfWeek: 1=Monday through 7=Sunday

Respond with a JSON object: {"workouts": [...]}
Each workout: {"runType": string, "week": int, "dayOfWeek": int, "description": string, "distance": number}
Distance is in kilometers as whole integers. Do not include any text outside the JSON object.`
}

func buildUserPrompt(input GenerateInput) string {
	return fmt.Sprintf(
		"Create a %d-week training plan with %d runs per week. Base weekly volume: %.1f km. "+
			"Distribute the volume across the runs with appropriate progression. "+
			"Remember: deload every 4th week, taper the last 3 weeks before race day (week %d).",
		input.Weeks, input.RunsPerWeek, input.BaseKmPerWeek, input.Weeks,
	)
}

type aiResponse struct {
	Workouts []BulkWorkoutInput `json:"workouts"`
}

func parseWorkouts(raw string) ([]BulkWorkoutInput, error) {
	var resp aiResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if len(resp.Workouts) == 0 {
		return nil, fmt.Errorf("no workouts in response")
	}
	for i := range resp.Workouts {
		resp.Workouts[i].Distance = math.Round(resp.Workouts[i].Distance)
	}
	return resp.Workouts, nil
}
