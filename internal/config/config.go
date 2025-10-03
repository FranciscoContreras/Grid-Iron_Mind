package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Environment       string
	DatabaseURL       string
	RedisURL          string
	APIKey            string
	UnlimitedAPIKey   string
	WeatherAPIKey     string
	YahooClientID     string
	YahooClientSecret string
	YahooRefreshToken string
	DBMaxConns        int32
	DBMinConns        int32
}

// LoadConfig reads configuration from environment variables (convenience wrapper)
func LoadConfig() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
	return cfg
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	cfg := &Config{
		Environment:       getEnv("ENVIRONMENT", "development"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		RedisURL:          getEnv("REDIS_URL", ""),
		APIKey:            getEnv("API_KEY", ""),
		UnlimitedAPIKey:   getEnv("UNLIMITED_API_KEY", ""),
		WeatherAPIKey:     getEnv("WEATHER_API_KEY", ""),
		YahooClientID:     getEnv("YAHOO_CLIENT_ID", ""),
		YahooClientSecret: getEnv("YAHOO_CLIENT_SECRET", ""),
		YahooRefreshToken: getEnv("YAHOO_REFRESH_TOKEN", ""),
		DBMaxConns:        int32(getEnvInt("DB_MAX_CONNS", 25)),
		DBMinConns:        int32(getEnvInt("DB_MIN_CONNS", 5)),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt reads an environment variable as int or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}