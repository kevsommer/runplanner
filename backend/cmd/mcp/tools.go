package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/service"
)

// runTypes are the workout types the backend accepts (see
// service.isValidRunType). "race" is created automatically from a plan's
// raceGoal, so it is not offered here.
var runTypes = []string{"easy_run", "intervals", "long_run", "tempo_run", "strength_training"}

var raceGoals = []string{"5k", "10k", "halfmarathon", "marathon"}

// statuses are the workout statuses the backend accepts (see
// service.WorkoutService.Update).
var statuses = []string{"pending", "completed", "skipped"}

func registerTools(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_training_plans",
		Title:       "List training plans",
		Description: "List all of the user's training plans with their id, name, race date, length in weeks and planned vs. completed kilometers. Use this to find the plan id needed by the other tools.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, listTrainingPlans(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_training_plan",
		Title:       "View a training plan",
		Description: "View one training plan in full: its metadata plus a week-by-week, day-by-day breakdown of every workout. Omit planId to view the user's currently active plan.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, getTrainingPlan(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:  "create_training_plan",
		Title: "Create a training plan",
		Description: "Create an empty training plan. The plan ends on the race date and spans the given number of weeks; week 1 starts on the Monday that many weeks before race week. " +
			"Passing raceGoal also creates the race-day workout. Add the training workouts afterwards with create_workouts.",
	}, createTrainingPlan(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:  "create_workouts",
		Title: "Create workouts in a plan",
		Description: "Add one or more workouts to a training plan, positioned by week number and day of week. This is the usual way to fill a plan: send the whole week (or the whole plan) in a single call. " +
			"Nothing is created if any workout in the batch is invalid.",
	}, createWorkouts(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_workout",
		Title:       "Create a single workout on a date",
		Description: "Add a single workout to a training plan on a specific calendar date. Prefer create_workouts when adding several workouts or when positioning them by training week.",
	}, createWorkout(c))

	mcp.AddTool(s, &mcp.Tool{
		Name:  "update_workout",
		Title: "Edit a workout",
		Description: "Edit an existing workout: change its run type, date, description, distance, notes or status (pending, completed, skipped). " +
			"Only the fields you pass are changed. Workout ids come from get_training_plan.",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, updateWorkout(c))
}

type listPlansInput struct {
	IncludeArchived bool `json:"includeArchived,omitempty" jsonschema:"include archived plans in the list (default false)"`
}

func listTrainingPlans(c *apiClient) mcp.ToolHandlerFor[listPlansInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listPlansInput) (*mcp.CallToolResult, any, error) {
		var resp struct {
			Plans []*service.PlanSummary `json:"plans"`
		}
		if err := c.do(ctx, "GET", "/api/plans", nil, &resp); err != nil {
			return toolError(err), nil, nil
		}

		plans := resp.Plans
		if !in.IncludeArchived {
			kept := make([]*service.PlanSummary, 0, len(plans))
			for _, p := range plans {
				if p.ArchivedAt == nil {
					kept = append(kept, p)
				}
			}
			plans = kept
		}

		activeID, err := c.activePlanID(ctx)
		if err != nil {
			return toolError(err), nil, nil
		}

		if len(plans) == 0 {
			return textResult("No training plans yet. Create one with create_training_plan."), nil, nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "%d training plan(s):\n", len(plans))
		for _, p := range plans {
			fmt.Fprintf(&b, "\n- %s (id: %s)\n", p.Name, p.ID)
			fmt.Fprintf(&b, "  %d weeks, %s to race day %s\n", p.Weeks, p.StartDate.Format(dateLayout), p.EndDate.Format(dateLayout))
			fmt.Fprintf(&b, "  %.1f km planned, %.1f km completed\n", p.TotalPlannedKm, p.TotalDoneKm)
			var flags []string
			if activeID != nil && *activeID == p.ID {
				flags = append(flags, "active")
			}
			if p.ArchivedAt != nil {
				flags = append(flags, "archived "+p.ArchivedAt.Format(dateLayout))
			}
			if len(flags) > 0 {
				fmt.Fprintf(&b, "  %s\n", strings.Join(flags, ", "))
			}
		}
		return textResult(b.String()), nil, nil
	}
}

type getPlanInput struct {
	PlanID string `json:"planId,omitempty" jsonschema:"id of the plan to view; omit to use the user's active plan"`
	Week   int    `json:"week,omitempty" jsonschema:"show only this training week (1-based); omit to show every week"`
}

func getTrainingPlan(c *apiClient) mcp.ToolHandlerFor[getPlanInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getPlanInput) (*mcp.CallToolResult, any, error) {
		planID, err := c.resolvePlanID(ctx, in.PlanID)
		if err != nil {
			return toolError(err), nil, nil
		}

		plan, err := c.getPlan(ctx, planID)
		if err != nil {
			return toolError(err), nil, nil
		}
		if in.Week != 0 && (in.Week < 1 || in.Week > plan.Weeks) {
			return toolError(fmt.Errorf("week must be between 1 and %d for this plan", plan.Weeks)), nil, nil
		}

		return textResult(formatPlanDetail(plan, in.Week)), nil, nil
	}
}

func formatPlanDetail(plan *service.PlanDetail, onlyWeek int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (id: %s)\n", plan.Name, plan.ID)
	fmt.Fprintf(&b, "%d weeks, %s to race day %s", plan.Weeks, plan.StartDate.Format(dateLayout), plan.EndDate.Format(dateLayout))
	if plan.ArchivedAt != nil {
		fmt.Fprintf(&b, " (archived)")
	}
	b.WriteString("\n")

	for _, week := range plan.WeeksSummary {
		if onlyWeek != 0 && week.Number != onlyWeek {
			continue
		}
		fmt.Fprintf(&b, "\nWeek %d — %.1f km planned, %.1f km done", week.Number, week.PlannedKm, week.DoneKm)
		if week.AllDone {
			b.WriteString(" (complete)")
		}
		b.WriteString("\n")
		for _, day := range week.Days {
			if len(day.Workouts) == 0 {
				fmt.Fprintf(&b, "  %s %s: rest\n", day.DayName, day.Date)
				continue
			}
			for _, w := range day.Workouts {
				fmt.Fprintf(&b, "  %s %s: %s", day.DayName, day.Date, w.RunType)
				if w.Distance > 0 {
					fmt.Fprintf(&b, ", %.1f km", w.Distance)
				}
				fmt.Fprintf(&b, " [%s] (id: %s)\n", w.Status, w.ID)
				if w.Description != "" {
					fmt.Fprintf(&b, "      %s\n", w.Description)
				}
				if w.Notes != "" {
					fmt.Fprintf(&b, "      notes: %s\n", w.Notes)
				}
			}
		}
	}
	return b.String()
}

type createPlanInput struct {
	Name     string `json:"name" jsonschema:"name of the training plan"`
	EndDate  string `json:"endDate" jsonschema:"race day as YYYY-MM-DD; the plan's last week ends here"`
	Weeks    int    `json:"weeks" jsonschema:"how many weeks the plan spans, at least 1"`
	RaceGoal string `json:"raceGoal,omitempty" jsonschema:"race distance - one of 5k, 10k, halfmarathon, marathon; when given, the race-day workout is created too"`
}

func createTrainingPlan(c *apiClient) mcp.ToolHandlerFor[createPlanInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in createPlanInput) (*mcp.CallToolResult, any, error) {
		if in.RaceGoal != "" && !contains(raceGoals, in.RaceGoal) {
			return toolError(fmt.Errorf("raceGoal must be one of %s", strings.Join(raceGoals, ", "))), nil, nil
		}

		var resp struct {
			Plan *model.TrainingPlan `json:"plan"`
		}
		if err := c.do(ctx, "POST", "/api/plans", in, &resp); err != nil {
			return toolError(err), nil, nil
		}

		plan := resp.Plan
		return textResult(fmt.Sprintf(
			"Created training plan %q (id: %s).\n%d weeks, week 1 starts %s, race day %s.\nAdd workouts with create_workouts using this plan id.",
			plan.Name, plan.ID, plan.Weeks, plan.StartDate.Format(dateLayout), plan.EndDate.Format(dateLayout),
		)), nil, nil
	}
}

type workoutItem struct {
	RunType     string  `json:"runType" jsonschema:"one of easy_run, intervals, long_run, tempo_run, strength_training"`
	Week        int     `json:"week" jsonschema:"training week the workout belongs to, 1-based"`
	DayOfWeek   int     `json:"dayOfWeek" jsonschema:"day within the week: 1 is Monday through 7 is Sunday"`
	Description string  `json:"description,omitempty" jsonschema:"what the session consists of, e.g. 6x800m at 5k pace with 2min jog"`
	Distance    float64 `json:"distance,omitempty" jsonschema:"distance in kilometers; must be 0 for strength_training"`
}

type createWorkoutsInput struct {
	PlanID   string        `json:"planId,omitempty" jsonschema:"id of the plan to add the workouts to; omit to use the user's active plan"`
	Workouts []workoutItem `json:"workouts" jsonschema:"the workouts to create"`
}

func createWorkouts(c *apiClient) mcp.ToolHandlerFor[createWorkoutsInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in createWorkoutsInput) (*mcp.CallToolResult, any, error) {
		if len(in.Workouts) == 0 {
			return toolError(fmt.Errorf("workouts must contain at least one workout")), nil, nil
		}
		for i, w := range in.Workouts {
			if err := validateRunType(w.RunType, w.Distance); err != nil {
				return toolError(fmt.Errorf("workouts[%d]: %w", i, err)), nil, nil
			}
		}

		planID, err := c.resolvePlanID(ctx, in.PlanID)
		if err != nil {
			return toolError(err), nil, nil
		}

		var resp struct {
			Workouts []*model.Workout `json:"workouts"`
		}
		body := map[string]any{"workouts": in.Workouts}
		if err := c.do(ctx, "POST", "/api/plans/"+url.PathEscape(planID)+"/workouts/bulk", body, &resp); err != nil {
			return toolError(err), nil, nil
		}

		var b strings.Builder
		var totalKm float64
		fmt.Fprintf(&b, "Created %d workout(s) in plan %s:\n", len(resp.Workouts), planID)
		for _, w := range resp.Workouts {
			totalKm += w.Distance
			fmt.Fprintf(&b, "- %s %s", w.Day.Format(dateLayout), w.RunType)
			if w.Distance > 0 {
				fmt.Fprintf(&b, ", %.1f km", w.Distance)
			}
			fmt.Fprintf(&b, " (id: %s)\n", w.ID)
		}
		fmt.Fprintf(&b, "%.1f km total.", totalKm)
		return textResult(b.String()), nil, nil
	}
}

type createWorkoutInput struct {
	PlanID      string  `json:"planId,omitempty" jsonschema:"id of the plan to add the workout to; omit to use the user's active plan"`
	RunType     string  `json:"runType" jsonschema:"one of easy_run, intervals, long_run, tempo_run, strength_training"`
	Day         string  `json:"day" jsonschema:"calendar date of the workout as YYYY-MM-DD; must fall inside the plan"`
	Description string  `json:"description,omitempty" jsonschema:"what the session consists of"`
	Distance    float64 `json:"distance,omitempty" jsonschema:"distance in kilometers; must be 0 for strength_training"`
}

func createWorkout(c *apiClient) mcp.ToolHandlerFor[createWorkoutInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in createWorkoutInput) (*mcp.CallToolResult, any, error) {
		if err := validateRunType(in.RunType, in.Distance); err != nil {
			return toolError(err), nil, nil
		}

		day, err := time.Parse(dateLayout, in.Day)
		if err != nil {
			return toolError(fmt.Errorf("day must be a date of the form YYYY-MM-DD")), nil, nil
		}

		planID, err := c.resolvePlanID(ctx, in.PlanID)
		if err != nil {
			return toolError(err), nil, nil
		}

		// The API happily stores a workout dated outside the plan, where the
		// plan view would never show it. Catch that here instead.
		plan, err := c.getPlan(ctx, planID)
		if err != nil {
			return toolError(err), nil, nil
		}
		lastDay := plan.StartDate.AddDate(0, 0, plan.Weeks*7-1)
		if day.Before(plan.StartDate) || day.After(lastDay) {
			return toolError(fmt.Errorf("day %s falls outside plan %q, which runs %s to %s",
				in.Day, plan.Name, plan.StartDate.Format(dateLayout), lastDay.Format(dateLayout))), nil, nil
		}

		body := map[string]any{
			"planId":      planID,
			"runType":     in.RunType,
			"day":         in.Day,
			"description": in.Description,
			"distance":    in.Distance,
		}
		var resp struct {
			Workout *model.Workout `json:"workout"`
		}
		if err := c.do(ctx, "POST", "/api/workouts", body, &resp); err != nil {
			return toolError(err), nil, nil
		}

		w := resp.Workout
		out := fmt.Sprintf("Created %s on %s in plan %s (id: %s)", w.RunType, w.Day.Format(dateLayout), planID, w.ID)
		if w.Distance > 0 {
			out = fmt.Sprintf("Created %s of %.1f km on %s in plan %s (id: %s)", w.RunType, w.Distance, w.Day.Format(dateLayout), planID, w.ID)
		}
		return textResult(out), nil, nil
	}
}

func validateRunType(runType string, distance float64) error {
	if !contains(runTypes, runType) {
		return fmt.Errorf("runType %q is not valid, must be one of %s", runType, strings.Join(runTypes, ", "))
	}
	if runType == "strength_training" && distance != 0 {
		return fmt.Errorf("strength_training must have a distance of 0")
	}
	if distance < 0 {
		return fmt.Errorf("distance cannot be negative")
	}
	return nil
}

// resolvePlanID falls back to the user's active plan when no id is given, so
// the common "add a workout to what I'm training for" case needs no lookup.
func (c *apiClient) resolvePlanID(ctx context.Context, planID string) (string, error) {
	if planID != "" {
		return planID, nil
	}
	active, err := c.activePlanID(ctx)
	if err != nil {
		return "", err
	}
	if active == nil {
		return "", fmt.Errorf("no planId given and the user has no active plan; pass planId (see list_training_plans)")
	}
	return string(*active), nil
}

func (c *apiClient) getPlan(ctx context.Context, planID string) (*service.PlanDetail, error) {
	var resp struct {
		Plan *service.PlanDetail `json:"plan"`
	}
	if err := c.do(ctx, "GET", "/api/plans/"+url.PathEscape(planID), nil, &resp); err != nil {
		return nil, err
	}
	if resp.Plan == nil {
		return nil, fmt.Errorf("plan %s not found", planID)
	}
	return resp.Plan, nil
}

func (c *apiClient) activePlanID(ctx context.Context) (*model.TrainingPlanID, error) {
	var resp struct {
		User model.PublicUser `json:"user"`
	}
	if err := c.do(ctx, "GET", "/api/auth/me", nil, &resp); err != nil {
		return nil, err
	}
	return resp.User.ActivePlanID, nil
}

func contains(values []string, v string) bool {
	for _, candidate := range values {
		if candidate == v {
			return true
		}
	}
	return false
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

// toolError reports a failure to the model rather than to the protocol, so the
// model can read the message and correct its call.
func toolError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}
}

type updateWorkoutInput struct {
	WorkoutID   string   `json:"workoutId" jsonschema:"id of the workout to edit, as shown by get_training_plan"`
	RunType     *string  `json:"runType,omitempty" jsonschema:"new run type - one of easy_run, intervals, long_run, tempo_run, strength_training"`
	Day         *string  `json:"day,omitempty" jsonschema:"move the workout to this calendar date, as YYYY-MM-DD; must stay inside the plan"`
	Description *string  `json:"description,omitempty" jsonschema:"new description of the session; pass an empty string to clear it"`
	Notes       *string  `json:"notes,omitempty" jsonschema:"new notes on how the session went; pass an empty string to clear them"`
	Status      *string  `json:"status,omitempty" jsonschema:"new status - one of pending, completed, skipped"`
	Distance    *float64 `json:"distance,omitempty" jsonschema:"new distance in kilometers; must be 0 for strength_training"`
}

func updateWorkout(c *apiClient) mcp.ToolHandlerFor[updateWorkoutInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in updateWorkoutInput) (*mcp.CallToolResult, any, error) {
		if in.WorkoutID == "" {
			return toolError(fmt.Errorf("workoutId is required; get_training_plan lists the id of every workout")), nil, nil
		}
		if in.RunType == nil && in.Day == nil && in.Description == nil && in.Notes == nil && in.Status == nil && in.Distance == nil {
			return toolError(fmt.Errorf("pass at least one field to change")), nil, nil
		}
		if in.Status != nil && !contains(statuses, *in.Status) {
			return toolError(fmt.Errorf("status %q is not valid, must be one of %s", *in.Status, strings.Join(statuses, ", "))), nil, nil
		}

		before, err := c.getWorkout(ctx, in.WorkoutID)
		if err != nil {
			return toolError(err), nil, nil
		}

		// Validate the workout as it will be once the changes are applied: a
		// distance-only edit still has to agree with the run type it keeps.
		// The run type itself is only checked when it is being set, since a
		// race workout legitimately carries a type create_workout never offers.
		runType, distance := before.RunType, before.Distance
		if in.RunType != nil {
			runType = *in.RunType
		}
		if in.Distance != nil {
			distance = *in.Distance
		}
		if in.RunType != nil && !contains(runTypes, runType) {
			return toolError(fmt.Errorf("runType %q is not valid, must be one of %s", runType, strings.Join(runTypes, ", "))), nil, nil
		}
		if in.RunType != nil || in.Distance != nil {
			if distance < 0 {
				return toolError(fmt.Errorf("distance cannot be negative")), nil, nil
			}
			if runType == "strength_training" && distance != 0 {
				return toolError(fmt.Errorf("strength_training must have a distance of 0")), nil, nil
			}
		}

		if in.Day != nil {
			day, err := time.Parse(dateLayout, *in.Day)
			if err != nil {
				return toolError(fmt.Errorf("day must be a date of the form YYYY-MM-DD")), nil, nil
			}
			plan, err := c.getPlan(ctx, string(before.PlanID))
			if err != nil {
				return toolError(err), nil, nil
			}
			// As in create_workout: the API stores an out-of-range date that
			// the plan view would then never show.
			lastDay := plan.StartDate.AddDate(0, 0, plan.Weeks*7-1)
			if day.Before(plan.StartDate) || day.After(lastDay) {
				return toolError(fmt.Errorf("day %s falls outside plan %q, which runs %s to %s",
					*in.Day, plan.Name, plan.StartDate.Format(dateLayout), lastDay.Format(dateLayout))), nil, nil
			}
		}

		var resp struct {
			Workout *model.Workout `json:"workout"`
		}
		if err := c.do(ctx, "PUT", "/api/workouts/"+url.PathEscape(in.WorkoutID), in, &resp); err != nil {
			return toolError(err), nil, nil
		}

		return textResult(formatWorkoutUpdate(before, resp.Workout)), nil, nil
	}
}

// formatWorkoutUpdate reports only the fields that actually changed, so the
// model can see at a glance whether the edit did what it asked for.
func formatWorkoutUpdate(before, after *model.Workout) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Updated workout %s in plan %s.\n", after.ID, after.PlanID)

	changes := []struct{ field, from, to string }{
		{"runType", before.RunType, after.RunType},
		{"day", before.Day.Format(dateLayout), after.Day.Format(dateLayout)},
		{"distance", fmt.Sprintf("%.1f km", before.Distance), fmt.Sprintf("%.1f km", after.Distance)},
		{"status", before.Status, after.Status},
		{"description", before.Description, after.Description},
		{"notes", before.Notes, after.Notes},
	}
	changed := false
	for _, ch := range changes {
		if ch.from == ch.to {
			continue
		}
		changed = true
		fmt.Fprintf(&b, "- %s: %s -> %s\n", ch.field, quoteIfEmpty(ch.from), quoteIfEmpty(ch.to))
	}
	if !changed {
		b.WriteString("No field changed; the workout already had these values.\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func quoteIfEmpty(v string) string {
	if v == "" {
		return `""`
	}
	return v
}

func (c *apiClient) getWorkout(ctx context.Context, workoutID string) (*model.Workout, error) {
	var resp struct {
		Workout *model.Workout `json:"workout"`
	}
	if err := c.do(ctx, "GET", "/api/workouts/"+url.PathEscape(workoutID), nil, &resp); err != nil {
		return nil, err
	}
	if resp.Workout == nil {
		return nil, fmt.Errorf("workout %s not found", workoutID)
	}
	return resp.Workout, nil
}
