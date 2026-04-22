package main

import (
	"context"
	"net/http"
	"time"

	"github.com/emekachisom/kolowise/internal/accounts"
	"github.com/emekachisom/kolowise/internal/auth"
	"github.com/emekachisom/kolowise/internal/goals"
	"github.com/emekachisom/kolowise/internal/mlbridge"
	"github.com/emekachisom/kolowise/internal/mlclient"
	"github.com/emekachisom/kolowise/internal/recommendations"
	"github.com/emekachisom/kolowise/internal/transactions"
	"github.com/emekachisom/kolowise/pkg/config"
	"github.com/emekachisom/kolowise/pkg/db"
	"github.com/emekachisom/kolowise/pkg/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	postgresPool, err := db.NewPostgres(cfg)
	if err != nil {
		panic(err)
	}
	defer postgresPool.Close()

	redisClient, err := db.NewRedis(cfg)
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	mlClient := mlclient.NewClient(cfg.MLServiceURL)

	authHandler := auth.NewHandler(postgresPool, cfg.JWTSecret, cfg.JWTIssuer)
	accountHandler := accounts.NewHandler(postgresPool)
	transactionHandler := transactions.NewHandler(postgresPool, mlClient)
	goalHandler := goals.NewHandler(postgresPool)
	recommendationHandler := recommendations.NewHandler(postgresPool, mlClient)
	mlBridgeHandler := mlbridge.NewHandler(mlClient)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		dbStatus := "ok"
		redisStatus := "ok"

		if err := postgresPool.Ping(ctx); err != nil {
			dbStatus = "down"
		}

		if err := redisClient.Ping(ctx).Err(); err != nil {
			redisStatus = "down"
		}

		statusCode := http.StatusOK
		if dbStatus != "ok" || redisStatus != "ok" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"service": "kolowise-api",
			"env":     cfg.AppEnv,
			"status": map[string]string{
				"api":      "ok",
				"postgres": dbStatus,
				"redis":    redisStatus,
			},
			"time": time.Now().UTC(),
		})
	})

	api := router.Group("/api/v1")

	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.GET("/me", middleware.AuthRequired(cfg.JWTSecret, cfg.JWTIssuer), authHandler.Me)
	}

	protected := api.Group("/")
	protected.Use(middleware.AuthRequired(cfg.JWTSecret, cfg.JWTIssuer))
	{
		protected.POST("/accounts", accountHandler.Create)
		protected.GET("/accounts", accountHandler.List)

		protected.POST("/transactions/manual", transactionHandler.CreateManual)
		protected.POST("/transactions/upload-csv", transactionHandler.UploadCSV)
		protected.GET("/transactions", transactionHandler.List)

		protected.POST("/goals", goalHandler.Create)
		protected.GET("/goals", goalHandler.List)
		protected.POST("/goals/:id/contribute", goalHandler.Contribute)

		protected.GET("/insights/safe-to-save", recommendationHandler.SafeToSave)

		protected.POST("/ml/predict-category", mlBridgeHandler.PredictCategory)
		protected.POST("/ml/predict-safe-to-save", mlBridgeHandler.PredictSafeToSave)
	}

	if err := router.Run(":" + cfg.APIPort); err != nil {
		panic(err)
	}
}
