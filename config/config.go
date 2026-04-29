package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port             string
	DSN              string
	SessionSecret    string // hex-encoded; validated/decoded in main
	AdminSecret      string // hex-encoded; validated/decoded in main
	AllowedOrigins   string // comma-separated WS origin allowlist
	SeedRestaurantID string // optional: UUID of restaurant to seed initial admin for
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", "postgres"),
			getEnv("DB_NAME", "waiter"),
		)
	}

	return &Config{
		Port:             port,
		DSN:              dsn,
		SessionSecret:    os.Getenv("SESSION_SECRET"),
		AdminSecret:      os.Getenv("ADMIN_SECRET"),
		AllowedOrigins:   os.Getenv("ALLOWED_ORIGINS"),
		SeedRestaurantID: os.Getenv("SEED_ADMIN_RESTAURANT_ID"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
