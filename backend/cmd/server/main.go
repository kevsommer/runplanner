package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/kevsommer/runplanner/internal/controller"
	"github.com/kevsommer/runplanner/internal/service"
	"github.com/kevsommer/runplanner/internal/store/mem"
)

func main() {
	port := getenv("PORT", "8080")
	sessionSecret := getenv("SESSION_SECRET", "dev-secret-change-me")

	// Wire dependencies
	userStore := mem.NewMemUserStore()
	authSvc := service.NewAuthService(userStore)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // set to false if you are NOT using cookies/auth
		MaxAge:           12 * time.Hour,
	}))

	// Sessions middleware
	storeCookie := cookie.NewStore([]byte(sessionSecret))
	storeCookie.Options(sessions.Options{ // secure defaults for local dev
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("rp.sid", storeCookie))

	// Health
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// API routes
	api := r.Group("/api")
	controller.RegisterAuthRoutes(api, authSvc)

	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
