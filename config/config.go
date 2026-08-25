package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	AppPort       string
	DatabaseURL   string
	JWTSecret     string
	JWTExpiresIn  string
	BasicAuthUser string
	BasicAuthPass string
	AdminName     string
	AdminEmail    string
	AdminPassword string
}

func Load() Config {
	// A missing .env file is valid in production, where configuration is
	// normally supplied by the runtime environment.
	_ = godotenv.Load()

	return Config{
		AppEnv:        valueOrDefault("APP_ENV", "development"),
		AppPort:       valueOrDefault("APP_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTExpiresIn:  valueOrDefault("JWT_EXPIRES_IN", "24h"),
		BasicAuthUser: os.Getenv("BASIC_AUTH_USERNAME"),
		BasicAuthPass: os.Getenv("BASIC_AUTH_PASSWORD"),
		AdminName:     os.Getenv("ADMIN_NAME"),
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	}
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
