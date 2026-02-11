package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/service"
	"github.com/kevsommer/runplanner/internal/store"
)

type WorkoutController struct {
	workouts *service.WorkoutService
	plans    *service.TrainingPlanService
}

func RegisterWorkoutRoutes(rg *gin.RouterGroup, workouts *service.WorkoutService, plans *service.TrainingPlanService) {
	wc := &WorkoutController{
		workouts: workouts,
		plans:    plans,
	}

	ws := rg.Group("/workouts")
	ws.Use(requireAuth)
	{
		ws.POST("/", wc.postCreate)
		ws.GET("/:id", wc.getByID)
	}

	plansGroup := rg.Group("/plans")
	plansGroup.Use(requireAuth)
	{
		plansGroup.GET("/:id/workouts", wc.getByPlanID)
	}
}

type createWorkoutInput struct {
	PlanID      string  `json:"planId" binding:"required"`
	RunType     string  `json:"runType" binding:"required"`
	Day         string  `json:"day" binding:"required"` // ISO date YYYY-MM-DD
	Description string  `json:"description" binding:"required"`
	Distance    float64 `json:"distance" binding:"required"`
}

func (w *WorkoutController) postCreate(c *gin.Context) {
	uid := currentUserID(c)

	var req createWorkoutInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "planId, runType, day, description and distance are required"})
		return
	}

	day, err := time.Parse("2006-01-02", req.Day)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "day must be YYYY-MM-DD"})
		return
	}

	plan, err := w.plans.GetByID(model.TrainingPlanID(req.PlanID))
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

	workout, err := w.workouts.Create(plan.ID, req.RunType, day, req.Description, req.Distance)
	if err != nil {
		switch err {
		case service.ErrInvalidDistance:
			c.JSON(http.StatusBadRequest, gin.H{"error": "distance cannot be negative"})
			return
		case service.ErrInvalidRunType:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid run type"})
			return
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"workout": workout})
}

func (w *WorkoutController) getByID(c *gin.Context) {
	uid := currentUserID(c)
	id := model.WorkoutID(c.Param("id"))

	workout, err := w.workouts.GetByID(id)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workout"})
		return
	}

	plan, err := w.plans.GetByID(workout.PlanID)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan"})
		return
	}
	if plan.UserID != model.UserID(uid) {
		c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workout": workout})
}

func (w *WorkoutController) getByPlanID(c *gin.Context) {
	uid := currentUserID(c)
	planID := model.TrainingPlanID(c.Param("id"))

	plan, err := w.plans.GetByID(planID)
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

	workouts, err := w.workouts.GetByPlanID(planID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"workouts": workouts})
}

