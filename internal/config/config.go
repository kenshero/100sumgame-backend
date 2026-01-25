package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	Port         string
	DatabaseURL  string
	GeminiAPIKey string
	Environment  string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/sum100game?sslmode=disable"),
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		Environment:  getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
