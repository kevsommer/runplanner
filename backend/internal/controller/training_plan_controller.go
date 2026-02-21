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
		plans.DELETE("/:id", tc.deletePlan)
		plans.POST("/:id/activate", tc.postActivate)
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
	c.JSON(http.StatusOK, gin.H{"plans": plans})
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
	}

	if err := t.auth.SetActivePlan(uid, newActivePlanID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update active plan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"activePlanId": newActivePlanID})
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
