package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/store"
)

var (
	ErrInvalidWeeks = errors.New("weeks must be at least 1")
	ErrInvalidName  = errors.New("name is required")
)

type TrainingPlanService struct {
	plans store.TrainingPlanStore
}

func NewTrainingPlanService(plans store.TrainingPlanStore) *TrainingPlanService {
	return &TrainingPlanService{plans: plans}
}

func StartDateFor(endDate time.Time, weeks int) time.Time {
	weekday := endDate.Weekday()
	daysSinceMonday := int(weekday) - 1
	if weekday == time.Sunday {
		daysSinceMonday = 6
	}
	mondayOfRaceWeek := endDate.AddDate(0, 0, -daysSinceMonday)
	mondayOfWeek1 := mondayOfRaceWeek.AddDate(0, 0, -(weeks-1)*7)
	return time.Date(mondayOfWeek1.Year(), mondayOfWeek1.Month(), mondayOfWeek1.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *TrainingPlanService) Create(userID model.UserID, name string, endDate time.Time, weeks int) (*model.TrainingPlan, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if weeks < 1 {
		return nil, ErrInvalidWeeks
	}
	startDate := StartDateFor(endDate, weeks)
	plan := &model.TrainingPlan{
		ID:        model.TrainingPlanID(newPlanID()),
		UserID:    userID,
		Name:      name,
		EndDate:   endDate,
		Weeks:     weeks,
		StartDate: startDate,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.plans.Create(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *TrainingPlanService) GetByID(id model.TrainingPlanID) (*model.TrainingPlan, error) {
	return s.plans.GetByID(id)
}

type DayDetail struct {
	Date     string           `json:"date"`
	DayName  string           `json:"dayName"`
	Workouts []*model.Workout `json:"workouts"`
}

type WeekSummary struct {
	Number    int         `json:"number"`
	PlannedKm float64    `json:"plannedKm"`
	DoneKm    float64    `json:"doneKm"`
	AllDone   bool        `json:"allDone"`
	Days      []DayDetail `json:"days"`
}

type PlanDetail struct {
	ID           model.TrainingPlanID `json:"id"`
	UserID       model.UserID         `json:"userId"`
	Name         string               `json:"name"`
	EndDate      time.Time            `json:"endDate"`
	Weeks        int                  `json:"weeks"`
	StartDate    time.Time            `json:"startDate"`
	CreatedAt    time.Time            `json:"createdAt"`
	WeeksSummary []WeekSummary        `json:"weeksSummary"`
}

var dayNames = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

func BuildPlanDetail(plan *model.TrainingPlan, workouts []*model.Workout) *PlanDetail {
	weeksSummary := make([]WeekSummary, plan.Weeks)

	for weekIdx := 0; weekIdx < plan.Weeks; weekIdx++ {
		days := make([]DayDetail, 7)
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			date := plan.StartDate.AddDate(0, 0, weekIdx*7+dayIdx)
			dateStr := date.Format("2006-01-02")

			var dayWorkouts []*model.Workout
			for _, w := range workouts {
				if w.Day.Format("2006-01-02") == dateStr {
					dayWorkouts = append(dayWorkouts, w)
				}
			}
			if dayWorkouts == nil {
				dayWorkouts = []*model.Workout{}
			}

			days[dayIdx] = DayDetail{
				Date:     dateStr,
				DayName:  dayNames[dayIdx],
				Workouts: dayWorkouts,
			}
		}

		var plannedKm, doneKm float64
		var totalWorkouts int
		var allCompleted int
		for _, d := range days {
			for _, w := range d.Workouts {
				totalWorkouts++
				plannedKm += w.Distance
				if w.Status == "completed" {
					doneKm += w.Distance
					allCompleted++
				} else if w.Status == "skipped" {
					allCompleted++
				}
			}
		}

		weeksSummary[weekIdx] = WeekSummary{
			Number:    weekIdx + 1,
			PlannedKm: plannedKm,
			DoneKm:    doneKm,
			AllDone:   totalWorkouts > 0 && allCompleted == totalWorkouts,
			Days:      days,
		}
	}

	return &PlanDetail{
		ID:           plan.ID,
		UserID:       plan.UserID,
		Name:         plan.Name,
		EndDate:      plan.EndDate,
		Weeks:        plan.Weeks,
		StartDate:    plan.StartDate,
		CreatedAt:    plan.CreatedAt,
		WeeksSummary: weeksSummary,
	}
}

func (s *TrainingPlanService) Update(id model.TrainingPlanID, name string, endDate time.Time, weeks int) (*model.TrainingPlan, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if weeks < 1 {
		return nil, ErrInvalidWeeks
	}
	plan, err := s.plans.GetByID(id)
	if err != nil {
		return nil, err
	}
	plan.Name = name
	plan.EndDate = endDate
	plan.Weeks = weeks
	plan.StartDate = StartDateFor(endDate, weeks)
	if err := s.plans.Update(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *TrainingPlanService) Delete(id model.TrainingPlanID) error {
	return s.plans.Delete(id)
}

func (s *TrainingPlanService) GetByUserID(userID model.UserID) ([]*model.TrainingPlan, error) {
	return s.plans.GetByUserID(userID)
}

func newPlanID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
