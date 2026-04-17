package main

import (
	"context"
	"net/http"
	"time"

	"github.com/emekachisom/kolowise/internal/auth"
	"github.com/emekachisom/kolowise/pkg/config"
	"github.com/emekachisom/kolowise/pkg/db"
	"github.com/emekachisom/kolowise/pkg/middleware"
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

	authHandler := auth.NewHandler(postgresPool, cfg.JWTSecret, cfg.JWTIssuer)

	router := gin.Default()

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

	router.Run(":" + cfg.APIPort)
}
