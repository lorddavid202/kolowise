package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	APIPort       string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	JWTSecret     string
	JWTIssuer     string
	MLServiceURL  string
}

func Load() Config {
	_ = godotenv.Load()

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		log.Fatalf("invalid REDIS_DB: %v", err)
	}

	cfg := Config{
		AppEnv:        getEnv("APP_ENV", "dev"),
		APIPort:       getEnv("API_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTIssuer:     getEnv("JWT_ISSUER", "kolowise"),
		MLServiceURL:  getEnv("ML_SERVICE_URL", "http://localhost:8000"),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	if cfg.JWTIssuer == "" {
		log.Fatal("JWT_ISSUER is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
