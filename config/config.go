package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	AppPort            string
	CORSAllowedOrigins []string
	DatabaseURL        string
	JWTSecret          string
	JWTExpiresIn       string
	BasicAuthUser      string
	BasicAuthPass      string
	AdminName          string
	AdminEmail         string
	AdminPassword      string
}

func Load() Config {
	// A missing .env file is valid in production, where configuration is
	// normally supplied by the runtime environment.
	_ = godotenv.Load()

	return Config{
		AppEnv:  valueOrDefault("APP_ENV", "development"),
		AppPort: valueOrDefault("APP_PORT", "8080"),
		CORSAllowedOrigins: parseAllowedOrigins(
			valueOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		),
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

func parseAllowedOrigins(value string) []string {
	origins := make([]string, 0)
	seen := make(map[string]struct{})

	for _, candidate := range strings.Split(value, ",") {
		origin := strings.TrimSpace(candidate)
		if origin == "" || origin == "*" {
			continue
		}
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return []string{"http://localhost:5173"}
	}
	return origins
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
