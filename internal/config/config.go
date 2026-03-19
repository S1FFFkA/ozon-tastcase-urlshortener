package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	StorageMemory   = "memory"
	StoragePostgres = "postgres"

	defaultDBHost          = "localhost"
	defaultDBPort          = "5432"
	defaultDBUser          = "postgres"
	defaultDBPassword      = "postgres"
	defaultDBName          = "url_shortener"
	defaultDBSSLMode       = "disable"
	defaultShutdownTimeout = 10 * time.Second
	defaultStorage         = StoragePostgres
	defaultCleanupBatch    = 100
	defaultRetentionYears  = 3
	defaultCleanupHour     = 3
	defaultCleanupMinute   = 0
)

type Config struct {
	HTTPAddr        string
	BaseURL         string
	Storage         string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	CleanupBatch    int
	RetentionYears  int
	CleanupHour     int
	CleanupMinute   int
	ShutdownTimeout time.Duration
}

func FromEnv() (*Config, error) {
	cfg := &Config{
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		BaseURL:         strings.TrimRight(getEnv("BASE_URL", "http://localhost:8080"), "/"),
		Storage:         strings.ToLower(getEnv("STORAGE", defaultStorage)),
		DBHost:          getEnv("DB_HOST", defaultDBHost),
		DBPort:          getEnv("DB_PORT", defaultDBPort),
		DBUser:          getEnv("DB_USER", defaultDBUser),
		DBPassword:      getEnv("DB_PASSWORD", defaultDBPassword),
		DBName:          getEnv("DB_NAME", defaultDBName),
		DBSSLMode:       getEnv("DB_SSLMODE", defaultDBSSLMode),
		CleanupBatch:    defaultCleanupBatch,
		RetentionYears:  defaultRetentionYears,
		CleanupHour:     defaultCleanupHour,
		CleanupMinute:   defaultCleanupMinute,
		ShutdownTimeout: defaultShutdownTimeout,
	}

	if raw := strings.TrimSpace(os.Getenv("SHUTDOWN_TIMEOUT")); raw != "" {
		timeout, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
		}
		if timeout <= 0 {
			return nil, errors.New("SHUTDOWN_TIMEOUT must be greater than zero")
		}
		cfg.ShutdownTimeout = timeout
	}

	if raw := strings.TrimSpace(os.Getenv("CLEANUP_BATCH_SIZE")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return nil, errors.New("CLEANUP_BATCH_SIZE must be a positive integer")
		}
		cfg.CleanupBatch = value
	}
	if raw := strings.TrimSpace(os.Getenv("CLEANUP_RETENTION_YEARS")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return nil, errors.New("CLEANUP_RETENTION_YEARS must be a positive integer")
		}
		cfg.RetentionYears = value
	}
	if raw := strings.TrimSpace(os.Getenv("CLEANUP_HOUR")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 || value > 23 {
			return nil, errors.New("CLEANUP_HOUR must be in range [0,23]")
		}
		cfg.CleanupHour = value
	}
	if raw := strings.TrimSpace(os.Getenv("CLEANUP_MINUTE")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 || value > 59 {
			return nil, errors.New("CLEANUP_MINUTE must be in range [0,59]")
		}
		cfg.CleanupMinute = value
	}

	switch cfg.Storage {
	case StorageMemory:
	case StoragePostgres:
	default:
		return nil, fmt.Errorf("unsupported STORAGE value: %s", cfg.Storage)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
