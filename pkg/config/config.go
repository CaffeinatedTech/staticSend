package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	Port                string
	DatabasePath        string
	EmailHost          string
	EmailPort          int
	EmailUsername      string
	EmailPassword      string
	EmailFrom          string
	EmailUseTLS        bool
	TurnstilePublicKey string
	TurnstileSecretKey string
	JWTSecretKey       string
	RegistrationEnabled bool
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		DatabasePath:        getEnv("DATABASE_PATH", "./data/staticsend.db"),
		EmailHost:          getEnv("EMAIL_HOST", "localhost"),
		EmailPort:          getEnvAsInt("EMAIL_PORT", 587),
		EmailUsername:      getEnv("EMAIL_USERNAME", ""),
		EmailPassword:      getEnv("EMAIL_PASSWORD", ""),
		EmailFrom:          getEnv("EMAIL_FROM", "noreply@example.com"),
		EmailUseTLS:        getEnvAsBool("EMAIL_USE_TLS", true),
		TurnstilePublicKey: getEnv("TURNSTILE_PUBLIC_KEY", ""),
		TurnstileSecretKey: getEnv("TURNSTILE_SECRET_KEY", ""),
		JWTSecretKey:       getEnv("JWT_SECRET_KEY", "change-this-secret-key"),
		RegistrationEnabled: getEnvAsBool("REGISTRATION_ENABLED", true),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvAsBool gets an environment variable as boolean with a fallback value
func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return fallback
}
