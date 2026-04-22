package config

import (
	"fmt"
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
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}

	loadEnvFiles(appEnv)

	redisDB, err := mustEnvInt("REDIS_DB")
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		AppEnv:        appEnv,
		APIPort:       mustEnv("API_PORT"),
		DatabaseURL:   mustEnv("DATABASE_URL"),
		RedisAddr:     mustEnv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,
		JWTSecret:     mustEnv("JWT_SECRET"),
		JWTIssuer:     mustEnv("JWT_ISSUER"),
		MLServiceURL:  mustEnv("ML_SERVICE_URL"),
	}
}

func loadEnvFiles(appEnv string) {
	baseEnv := ".env"
	envFile := fmt.Sprintf(".env.%s", appEnv)

	_ = godotenv.Load(baseEnv)
	_ = godotenv.Overload(envFile)
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s is required", key)
	}
	return val
}

func mustEnvInt(key string) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return 0, fmt.Errorf("%s is required", key)
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", key)
	}

	return parsed, nil
}
