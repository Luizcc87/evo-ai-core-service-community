package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Core               CoreConfig
	JWT                JWTConfig
	DB                 DBConfig
	Evolution          EvolutionConfig
	EvoAuth            EvoAuthConfig
	AIProcessorService AIProcessorServiceConfig
	RateLimit          RateLimitConfig
	HealthCheck        HealthCheckConfig
}

type CoreConfig struct {
	Port          string
	EncryptionKey string // Fernet encryption key for API keys (must be 32 URL-safe base64-encoded bytes)
}

// JWTConfig for validating Evolution JWT tokens
type JWTConfig struct {
	SecretKey string // Same secret key used by Evolution for token validation
	Algorithm string // Algorithm used by Evolution (usually HS256)
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxIdleConns    string
	MaxOpenConns    string
	ConnMaxLifetime string
	ConnMaxIdleTime string
}

// EvolutionConfig holds configuration for Evolution integration (for agent bot creation)
type EvolutionConfig struct {
	BaseURL string
}

// EvoAuthConfig holds configuration for EvoAuth service integration (for authentication)
type EvoAuthConfig struct {
	BaseURL string
}

type AIProcessorServiceConfig struct {
	URL     string
	Version string // Optional: if you want to specify a version for the auth service
}

type RateLimitConfig struct {
	GlobalRPS       float64
	GlobalBurst     int
	ClientRPS       float64
	ClientBurst     int
	CleanupInterval time.Duration
}

type HealthCheckConfig struct {
	Timeout                 time.Duration
	ReadinessTimeout        time.Duration
	DBQuery                 string
	DBConnThresholdDegraded int
	DBWaitCountThreshold    int
	MemoryThresholdWarning  int
	MemoryThresholdCritical int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	return &Config{
		Core: CoreConfig{
			Port:          LoadEnvOrPanic("PORT"),
			EncryptionKey: LoadEnvOrPanic("ENCRYPTION_KEY"), // Shared with evo-ai-processor
		},
		JWT: JWTConfig{
			SecretKey: LoadEnvOrPanic("JWT_SECRET_KEY"), // Same secret as Evolution
			Algorithm: LoadEnvOrDefault("JWT_ALGORITHM", "HS256"),
		},
		DB: DBConfig{
			Host:            LoadEnvOrPanic("DB_HOST"),
			Port:            LoadEnvOrPanic("DB_PORT"),
			User:            LoadEnvOrPanic("DB_USER"),
			Password:        LoadEnvOrPanic("DB_PASSWORD"),
			DBName:          LoadEnvOrPanic("DB_NAME"),
			SSLMode:         LoadEnvOrPanic("DB_SSLMODE"),
			MaxIdleConns:    LoadEnvOrDefault("DB_MAX_IDLE_CONNS", "10"),
			MaxOpenConns:    LoadEnvOrDefault("DB_MAX_OPEN_CONNS", "100"),
			ConnMaxLifetime: LoadEnvOrDefault("DB_CONN_MAX_LIFETIME", "1h"),
			ConnMaxIdleTime: LoadEnvOrDefault("DB_CONN_MAX_IDLE_TIME", "30m"), // 30 minutes
		},
		Evolution: EvolutionConfig{
			BaseURL: LoadEnvOrDefault("EVOLUTION_BASE_URL", ""),
		},
		EvoAuth: EvoAuthConfig{
			BaseURL: LoadEnvOrPanic("EVO_AUTH_BASE_URL"),
		},
		AIProcessorService: AIProcessorServiceConfig{
			URL:     LoadEnvOrPanic("AI_PROCESSOR_URL"),
			Version: LoadEnvOrPanic("AI_PROCESSOR_VERSION"), // Optional: you can set a default version or load it from an env variable
		},
		RateLimit: RateLimitConfig{
			GlobalRPS:       LoadEnvOrDefaultFloat("RATE_LIMIT_GLOBAL_RPS", 1000),
			GlobalBurst:     LoadEnvOrDefaultInt("RATE_LIMIT_GLOBAL_BURST", 50),
			ClientRPS:       LoadEnvOrDefaultFloat("RATE_LIMIT_CLIENT_RPS", 100),
			ClientBurst:     LoadEnvOrDefaultInt("RATE_LIMIT_CLIENT_BURST", 10),
			CleanupInterval: LoadEnvOrDefaultDuration("RATE_LIMIT_CLEANUP_INTERVAL", 5*time.Minute),
		},
		HealthCheck: HealthCheckConfig{
			Timeout:                 LoadEnvOrDefaultDuration("HEALTH_CHECK_TIMEOUT", 5*time.Second),
			ReadinessTimeout:        LoadEnvOrDefaultDuration("READINESS_CHECK_TIMEOUT", 2*time.Second),
			DBQuery:                 LoadEnvOrDefault("DB_HEALTH_CHECK_QUERY", "SELECT 1"),
			DBConnThresholdDegraded: LoadEnvOrDefaultInt("DB_CONN_THRESHOLD_DEGRADED", 90),
			DBWaitCountThreshold:    LoadEnvOrDefaultInt("DB_WAIT_COUNT_THRESHOLD", 100),
			MemoryThresholdWarning:  LoadEnvOrDefaultInt("MEMORY_THRESHOLD_WARNING", 400),
			MemoryThresholdCritical: LoadEnvOrDefaultInt("MEMORY_THRESHOLD_CRITICAL", 500),
		},
	}, nil
}

func LoadEnvOrPanic(variable string) string {
	value := os.Getenv(variable)
	if value == "" {
		log.Fatalf("Environment variable %s is required", variable)
	}
	return strings.TrimSpace(value)
}

func LoadEnvOrDefault(variable string, defaultValue string) string {
	value := os.Getenv(variable)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

func LoadEnvOrDefaultInt(variable string, defaultValue int) int {
	value := os.Getenv(variable)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		log.Printf("Warning: Invalid integer value for %s: %s, using default: %d", variable, value, defaultValue)
		return defaultValue
	}
	return intValue
}

func LoadEnvOrDefaultFloat(variable string, defaultValue float64) float64 {
	value := os.Getenv(variable)
	if value == "" {
		return defaultValue
	}

	floatValue, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		log.Printf("Warning: Invalid float value for %s: %s, using default: %f", variable, value, defaultValue)
		return defaultValue
	}
	return floatValue
}

func LoadEnvOrDefaultDuration(variable string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(variable)
	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		log.Printf("Warning: Invalid duration value for %s: %s, using default: %s", variable, value, defaultValue.String())
		return defaultValue
	}
	return duration
}
