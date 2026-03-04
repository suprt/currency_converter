package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	//API
	APIKey     string
	APIBaseURL string

	//Server
	ServerHost        string
	ServerPort        string
	ServerTimeout     time.Duration
	ServerIdleTimeout time.Duration

	//Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisUse      bool

	//Rate limiter
	RPS                    int
	Burst                  int
	RateLimiterCleanup     time.Duration

	//Updater
	UpdateInterval time.Duration

	// In-memory cache
	InMemoryCleanupInterval time.Duration

	// Admin
	AdminAPIKey string

	// Logger
	LogLevel string
}

func Load() *Config {
	// Try to load .env file, but don't fail if it doesn't exist
	// Environment variables can be set directly in docker-compose or system
	if err := godotenv.Load(); err != nil {
		// Silently ignore - env vars may be set externally
	}

	return &Config{
		APIKey:     getEnv("API_KEY", ""),
		APIBaseURL: getEnv("API_URL", "https://currencyapi.net/api/v2/"),

		ServerHost:        getEnv("SERVER_HOST", "localhost"),
		ServerPort:        getEnv("SERVER_PORT", ":8080"),
		ServerTimeout:     getEnvDuration("SERVER_TIMEOUT", 15*time.Second),
		ServerIdleTimeout: getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		RedisUse:      getEnvBool("REDIS_USE", false),

		RPS:                getEnvInt("RPS", 10),
		Burst:              getEnvInt("BURST", 20),
		RateLimiterCleanup: getEnvDuration("RATE_LIMITER_CLEANUP", 1*time.Minute),

		UpdateInterval: getEnvDuration("UPDATE_INTERVAL", 1*time.Hour),

		InMemoryCleanupInterval: getEnvDuration("INMEMORY_CLEANUP_INTERVAL", 5*time.Minute),

		AdminAPIKey: getEnv("ADMIN_API_KEY", ""),

		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}

	}
	return defaultValue
}

func (c *Config) ServerAddr() string {
	return c.ServerHost + ":" + c.ServerPort
}
