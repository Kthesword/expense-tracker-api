package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kthesword/expense-tracker-api/internal/auth"
	"github.com/kthesword/expense-tracker-api/internal/config"
	"github.com/kthesword/expense-tracker-api/internal/handlers"
	"github.com/kthesword/expense-tracker-api/internal/middleware"
	"github.com/kthesword/expense-tracker-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{TranslateError: true})
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Expense{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL)
	authH := &handlers.AuthHandler{DB: db, Tokens: tokens}
	expH := &handlers.ExpenseHandler{DB: db}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	r.POST("/signup", authH.Signup)
	r.POST("/login", authH.Login)

	// Protected routes
	api := r.Group("/expenses", middleware.RequireAuth(tokens))
	{
		api.POST("", expH.Create)
		api.GET("", expH.List)
	}

	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
