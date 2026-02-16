package controller

import (
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
}

func requireAuth(c *gin.Context) {
	if currentUserID(c) == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	c.Next()
}

func RegisterTrainingPlanRoutes(rg *gin.RouterGroup, svc *service.TrainingPlanService, workouts *service.WorkoutService) {
	tc := &TrainingPlanController{svc: svc, workouts: workouts}
	plans := rg.Group("/plans")
	plans.Use(requireAuth)
	{
		plans.POST("/", tc.postCreate)
		plans.GET("/", tc.getByUserID)
		plans.GET("/:id", tc.getByID)
	}
}

type createPlanInput struct {
	Name    string `json:"name" binding:"required"`
	EndDate string `json:"endDate" binding:"required"` // ISO date YYYY-MM-DD
	Weeks   int    `json:"weeks" binding:"required"`
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

func (t *TrainingPlanController) getByUserID(c *gin.Context) {
	uid := currentUserID(c)
	plans, err := t.svc.GetByUserID(model.UserID(uid))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}
