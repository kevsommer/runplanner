package controller

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/kevsommer/runplanner/internal/model"
	"github.com/kevsommer/runplanner/internal/service"
	"github.com/kevsommer/runplanner/internal/store"
)

type AuthController struct {
	svc *service.AuthService
}

func RegisterAuthRoutes(rg *gin.RouterGroup, svc *service.AuthService) {
	ac := &AuthController{svc: svc}
	auth := rg.Group("/auth")
	{
		auth.POST("/register", ac.postRegister)
		auth.POST("/login", ac.postLogin)
		auth.POST("/logout", ac.postLogout)
		auth.GET("/me", ac.getMe)
	}
}

type credentials struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (a *AuthController) postRegister(c *gin.Context) {
	var req credentials
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}
	u, err := a.svc.Register(req.Email, req.Password)
	if err != nil {
		switch err {
		case store.ErrEmailTaken:
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Auto-login after register
	sess := sessions.Default(c)
	sess.Set("uid", string(u.ID))
	_ = sess.Save()
	c.JSON(http.StatusCreated, gin.H{"user": u.Public()})
}

func (a *AuthController) postLogin(c *gin.Context) {
	var req credentials
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}
	u, err := a.svc.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	sess := sessions.Default(c)
	sess.Set("uid", string(u.ID))
	_ = sess.Save()
	c.JSON(http.StatusOK, gin.H{"user": u.Public()})
}

func (a *AuthController) postLogout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear()
	_ = sess.Save()
	c.Status(http.StatusNoContent)
}

func (a *AuthController) getMe(c *gin.Context) {
	uid := currentUserID(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	u, err := a.svc.GetUser(model.UserID(uid))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": u.Public()})
}

func currentUserID(c *gin.Context) string {
	sess := sessions.Default(c)
	if v := sess.Get("uid"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
