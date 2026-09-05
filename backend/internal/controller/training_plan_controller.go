package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/service"
	"github.com/kevsommer/runplanner/internal/store"
)

type TrainingPlanController struct {
	svc      *service.TrainingPlanService
	workouts *service.WorkoutService
	generate *service.GenerateService
	auth     *service.AuthService
}

func requireAuth(c *gin.Context) {
	if currentUserID(c) == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	c.Next()
}

func RegisterTrainingPlanRoutes(rg *gin.RouterGroup, svc *service.TrainingPlanService, workouts *service.WorkoutService, generate *service.GenerateService, auth *service.AuthService) {
	tc := &TrainingPlanController{svc: svc, workouts: workouts, generate: generate, auth: auth}
	plans := rg.Group("/plans")
	plans.Use(requireAuth)
	{
		plans.POST("", tc.postCreate)
		plans.POST("/generate", tc.postGenerate)
		plans.GET("", tc.getByUserID)
		plans.GET("/:id", tc.getByID)
		plans.PUT("/:id", tc.putUpdate)
		plans.PATCH("/:id", tc.patchUpdate)
		plans.DELETE("/:id", tc.deletePlan)
		plans.POST("/:id/activate", tc.postActivate)
		plans.POST("/:id/archive", tc.postArchive)
	}
}

type createPlanInput struct {
	Name     string `json:"name" binding:"required"`
	EndDate  string `json:"endDate" binding:"required"` // ISO date YYYY-MM-DD
	Weeks    int    `json:"weeks" binding:"required"`
	RaceGoal string `json:"raceGoal"`
}

func (t *TrainingPlanController) postCreate(c *gin.Context) {
	uid := currentUserID(c)
	var req createPlanInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, endDate and weeks are required"})
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate must be YYYY-MM-DD"})
		return
	}
	plan, err := t.svc.Create(model.UserID(uid), req.Name, endDate, req.Weeks)
	if err != nil {
		switch err {
		case service.ErrInvalidName:
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		case service.ErrInvalidWeeks:
			c.JSON(http.StatusBadRequest, gin.H{"error": "weeks must be at least 1"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.RaceGoal != "" {
		if _, err := t.workouts.CreateRaceWorkout(plan, req.RaceGoal); err != nil {
			_ = t.svc.Delete(plan.ID)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"plan": plan})
}

func (t *TrainingPlanController) getByID(c *gin.Context) {
	uid := currentUserID(c)
	id := model.TrainingPlanID(c.Param("id"))
	plan, err := t.svc.GetByID(id)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan"})
		return
	}
	if plan.UserID != model.UserID(uid) {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}
	workouts, err := t.workouts.GetByPlanID(plan.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workouts"})
		return
	}
	detail := service.BuildPlanDetail(plan, workouts)
	c.JSON(http.StatusOK, gin.H{"plan": detail})
}

type updatePlanInput struct {
	Name    string `json:"name" binding:"required"`
	EndDate string `json:"endDate" binding:"required"`
	Weeks   int    `json:"weeks" binding:"required"`
}

func (t *TrainingPlanController) putUpdate(c *gin.Context) {
	uid := currentUserID(c)
	id := model.TrainingPlanID(c.Param("id"))

	plan, err := t.svc.GetByID(id)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan"})
		return
	}
	if plan.UserID != model.UserID(uid) {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	var req updatePlanInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, endDate and weeks are required"})
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate must be YYYY-MM-DD"})
		return
	}

	updated, err := t.svc.Update(id, req.Name, endDate, req.Weeks)
	if err != nil {
		switch err {
		case service.ErrInvalidName:
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		case service.ErrInvalidWeeks:
			c.JSON(http.StatusBadRequest, gin.H{"error": "weeks must be at least 1"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update plan"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"plan": updated})
}

type patchPlanInput struct {
	Name *string `json:"name"`
}

// patchUpdate applies a partial update to a plan. Only the name can be changed
// this way; use PUT to change the dates.
func (t *TrainingPlanController) patchUpdate(c *gin.Context) {
	uid := currentUserID(c)
	id := model.TrainingPlanID(c.Param("id"))

	plan, err := t.svc.GetByID(id)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan"})
		return
	}
	if plan.UserID != model.UserID(uid) {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	var req patchPlanInput
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	updated, err := t.svc.Rename(id, *req.Name)
	if err != nil {
		if err == service.ErrInvalidName {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update plan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plan": updated})
}

func (t *TrainingPlanController) deletePlan(c *gin.Context) {
	uid := currentUserID(c)
	id := model.TrainingPlanID(c.Param("id"))

	plan, err := t.svc.GetByID(id)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan"})
		return
	}
	if plan.UserID != model.UserID(uid) {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	if err := t.svc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete plan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (t *TrainingPlanController) getByUserID(c *gin.Context) {
	uid := currentUserID(c)
	plans, err := t.svc.GetByUserID(model.UserID(uid))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	summaries := make([]*service.PlanSummary, 0, len(plans))
	for _, plan := range plans {
		workouts, err := t.workouts.GetByPlanID(plan.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workouts"})
			return
		}
		summaries = append(summaries, service.BuildPlanSummary(plan, workouts))
	}
	c.JSON(http.StatusOK, gin.H{"plans": summaries})
}

type generatePlanInput struct {
	Name          string  `json:"name" binding:"required"`
	EndDate       string  `json:"endDate" binding:"required"`
	Weeks         int     `json:"weeks" binding:"required"`
	BaseKmPerWeek float64 `json:"baseKmPerWeek" binding:"required"`
	RunsPerWeek   int     `json:"runsPerWeek" binding:"required"`
	RaceGoal      string  `json:"raceGoal" binding:"required"`
}

func (t *TrainingPlanController) postActivate(c *gin.Context) {
	uid := model.UserID(currentUserID(c))
	id := model.TrainingPlanID(c.Param("id"))

	plan, err := t.svc.GetByID(id)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan"})
		return
	}
	if plan.UserID != uid {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	user, err := t.auth.GetUser(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// Toggle: if this plan is already active, deactivate it; otherwise activate it.
	var newActivePlanID *model.TrainingPlanID
	if user.ActivePlanID == nil || *user.ActivePlanID != id {
		newActivePlanID = &id
		// Picking a plan back up takes it out of the archive.
		if plan.ArchivedAt != nil {
			if _, err := t.svc.SetArchived(id, false); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unarchive plan"})
				return
			}
		}
	}

	if err := t.auth.SetActivePlan(uid, newActivePlanID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update active plan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"activePlanId": newActivePlanID})
}

type archivePlanInput struct {
	Archived *bool `json:"archived" binding:"required"`
}

func (t *TrainingPlanController) postArchive(c *gin.Context) {
	uid := model.UserID(currentUserID(c))
	id := model.TrainingPlanID(c.Param("id"))

	plan, err := t.svc.GetByID(id)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan"})
		return
	}
	if plan.UserID != uid {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	var req archivePlanInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "archived is required"})
		return
	}

	updated, err := t.svc.SetArchived(id, *req.Archived)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to archive plan"})
		return
	}

	// An archived plan should not stay the user's active plan.
	activePlanID, err := t.clearActivePlanIfArchived(uid, updated)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update active plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": updated, "activePlanId": activePlanID})
}

// clearActivePlanIfArchived drops the user's active plan when that plan has just
// been archived, and reports the active plan id the user is left with.
func (t *TrainingPlanController) clearActivePlanIfArchived(uid model.UserID, plan *model.TrainingPlan) (*model.TrainingPlanID, error) {
	user, err := t.auth.GetUser(uid)
	if err != nil {
		return nil, err
	}
	if plan.ArchivedAt == nil || user.ActivePlanID == nil || *user.ActivePlanID != plan.ID {
		return user.ActivePlanID, nil
	}
	if err := t.auth.SetActivePlan(uid, nil); err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *TrainingPlanController) postGenerate(c *gin.Context) {
	uid := currentUserID(c)
	var req generatePlanInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, endDate, weeks, baseKmPerWeek, runsPerWeek and raceGoal are required"})
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate must be YYYY-MM-DD"})
		return
	}

	input := service.GenerateInput{
		Name:          req.Name,
		EndDate:       endDate,
		Weeks:         req.Weeks,
		BaseKmPerWeek: req.BaseKmPerWeek,
		RunsPerWeek:   req.RunsPerWeek,
		RaceGoal:      req.RaceGoal,
	}

	plan, workouts, err := t.generate.Generate(c.Request.Context(), model.UserID(uid), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAINotConfigured):
			c.JSON(http.StatusBadRequest, gin.H{"error": "AI generation is not configured on the server"})
		case errors.Is(err, service.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrAIGeneration):
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate plan"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"plan": plan, "workouts": workouts})
}
