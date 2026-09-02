package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName      string
	ServiceVersion   string
	CommitSHA        string
	BuildTime        string
	Port             string
	DatabaseDSN      string
	JWTSecret        string
	InternalAPIToken string
	LogLevel         string
	RequestTimeout   time.Duration
	StorageDir       string
	AutoMigrate      bool
	MaxVideoBytes    int64
	MaxImageBytes    int64
	UserServiceURL   string
}

func Load() (Config, error) {
	cfg := Config{
		ServiceName:      env("SERVICE_NAME", "content-service"),
		ServiceVersion:   env("SERVICE_VERSION", "microservice-0.1.0"),
		CommitSHA:        env("COMMIT_SHA", "0000000"),
		BuildTime:        env("BUILD_TIME", "1970-01-01T00:00:00Z"),
		Port:             env("PORT", "8080"),
		DatabaseDSN:      os.Getenv("DATABASE_DSN"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		InternalAPIToken: os.Getenv("INTERNAL_API_TOKEN"),
		LogLevel:         env("LOG_LEVEL", "info"),
		StorageDir:       env("STORAGE_DIR", "./data"),
		AutoMigrate:      envBool("AUTO_MIGRATE", false),
		MaxVideoBytes:    envInt64("MAX_VIDEO_BYTES", 2<<30),
		MaxImageBytes:    envInt64("MAX_IMAGE_BYTES", 10<<20),
		UserServiceURL:   env("USER_SERVICE_URL", "http://user-service:8080"),
	}
	timeout, err := time.ParseDuration(env("REQUEST_TIMEOUT", "2s"))
	if err != nil || timeout <= 0 || timeout > 2*time.Second {
		return Config{}, fmt.Errorf("REQUEST_TIMEOUT must be between 1ns and 2s")
	}
	cfg.RequestTimeout = timeout
	if cfg.DatabaseDSN == "" || cfg.JWTSecret == "" || cfg.InternalAPIToken == "" {
		return Config{}, errors.New("DATABASE_DSN, JWT_SECRET and INTERNAL_API_TOKEN are required")
	}
	if cfg.ServiceName != "content-service" {
		return Config{}, errors.New("SERVICE_NAME must be content-service")
	}
	if len(cfg.CommitSHA) < 7 {
		return Config{}, errors.New("COMMIT_SHA must contain at least 7 characters")
	}
	if _, err := time.Parse(time.RFC3339, cfg.BuildTime); err != nil {
		return Config{}, errors.New("BUILD_TIME must be RFC3339")
	}
	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, errors.New("PORT must be between 1 and 65535")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
