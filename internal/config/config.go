package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Application
	AppName        string
	AppVersion     string
	Environment    string
	AllowedOrigins string

	// Database
	DatabaseURL string

	// Cache
	RedisURL string

	// Storage
	Storage StorageConfig

	// Messaging
	NatsURL string

	// Authentication
	SuperTokens SuperTokensConfig

	// Observability
	TracingEnabled bool
	MetricsEnabled bool
}

// StorageConfig holds MinIO storage configuration
type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

// SuperTokensConfig holds SuperTokens configuration
type SuperTokensConfig struct {
	ConnectionURI   string
	APIKey          string
	AppName         string
	APIDomain       string
	WebsiteDomain   string
	APIBasePath     string
	WebsiteBasePath string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		AppName:        getEnv("APP_NAME", "fowergram-backend"),
		AppVersion:     getEnv("APP_VERSION", "1.0.0"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://fowergram:password@localhost:5432/fowergram?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),

		Storage: StorageConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          getEnvBool("MINIO_USE_SSL", false),
			BucketName:      getEnv("MINIO_BUCKET", "fowergram"),
		},

		SuperTokens: SuperTokensConfig{
			ConnectionURI:   getEnv("SUPERTOKENS_CONNECTION_URI", "http://localhost:3567"),
			APIKey:          getEnv("SUPERTOKENS_API_KEY", ""),
			AppName:         getEnv("SUPERTOKENS_APP_NAME", "Fowergram"),
			APIDomain:       getEnv("SUPERTOKENS_API_DOMAIN", "http://localhost:8000"),
			WebsiteDomain:   getEnv("SUPERTOKENS_WEBSITE_DOMAIN", "http://localhost:3000"),
			APIBasePath:     getEnv("SUPERTOKENS_API_BASE_PATH", "/auth"),
			WebsiteBasePath: getEnv("SUPERTOKENS_WEBSITE_BASE_PATH", "/auth"),
		},

		TracingEnabled: getEnvBool("TRACING_ENABLED", true),
		MetricsEnabled: getEnvBool("METRICS_ENABLED", true),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvBool gets a boolean environment variable with a fallback value
func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return fallback
}
