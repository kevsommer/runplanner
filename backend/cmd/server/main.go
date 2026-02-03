package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	goose "github.com/pressly/goose/v3"

	"github.com/kevsommer/runplanner/internal/controller"
	"github.com/kevsommer/runplanner/internal/service"
	"github.com/kevsommer/runplanner/internal/store"
	"github.com/kevsommer/runplanner/internal/store/mem"
	sqliteStore "github.com/kevsommer/runplanner/internal/store/sqlite"
)

func main() {
	port := getenv("PORT", "8080")
	sessionSecret := getenv("SESSION_SECRET", "dev-secret-change-me")
	dbURL := getenv("DATABASE_URL", "") // e.g. file:data/runplanner.db?_pragma=busy_timeout(5000)&cache=shared

	// Choose store (SQLite if DATABASE_URL provided; else in-memory)
	var userStore store.UserStore
	var trainingPlanStore store.TrainingPlanStore
	if dbURL == "" {
		userStore = mem.NewMemUserStore()
		trainingPlanStore = mem.NewMemTrainingPlanStore()
	} else {
		db, err := sqliteStore.Open(dbURL) // uses modernc.org/sqlite
		if err != nil {
			log.Fatalf("open sqlite: %v", err)
		}
		if err := runMigrations(db); err != nil {
			log.Fatalf("migrations: %v", err)
		}
		userStore = sqliteStore.NewUserStore(db)
		trainingPlanStore = sqliteStore.NewTrainingPlanStore(db)
	}

	authSvc := service.NewAuthService(userStore)
	trainingPlanSvc := service.NewTrainingPlanService(trainingPlanStore)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Sessions middleware
	storeCookie := cookie.NewStore([]byte(sessionSecret))
	storeCookie.Options(sessions.Options{
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
	controller.RegisterTrainingPlanRoutes(api, trainingPlanSvc)

	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func runMigrations(db *sql.DB) error {
	// Reads migrations from the filesystem at ./db/migrations
	// Example migration file: db/migrations/0001_create_users.sql
	return goose.Up(db, "db/migrations")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
