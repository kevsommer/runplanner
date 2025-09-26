package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"time"

	"github.com/kevsommer/runplanner/internal/service"
)

type TrainingPlanController struct {
	svc *service.TrainingPlanService
}

func RegisterTrainingPlanRoutes(rg *gin.RouterGroup, svc *service.TrainingPlanService) {
	ac := &TrainingPlanController{svc: svc}
	auth := rg.Group("/plans")
	{
		auth.POST("/", ac.postCreate)
	}
}

type plan_credentials struct {
	Goal              string    `json:"goal" binding:"required"`
	StartDate         time.Time `json:"start_date" binding:"required"`
	NumberOfWeeks     int       `json:"number_of_weeks" binding:"required"`
	ActivitiesPerWeek int       `json:"activities_per_week" binding:"required"`
	Name              string    `json:"name" binding:"required"`
}

func (a *TrainingPlanController) postCreate(c *gin.Context) {
	var req plan_credentials
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "all fields are required"})
		return
	}

	uid := currentUserID(c)

	tp, err := a.svc.Create(uid, req.Goal, req.StartDate, req.NumberOfWeeks, req.ActivitiesPerWeek, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"plan": tp})
}
