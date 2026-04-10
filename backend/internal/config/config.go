package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName          string
	Environment      string
	HTTPPort         string
	DatabaseURL      string
	JWTAccessSecret  string
	JWTRefreshSecret string
	AccessTTL        time.Duration
	RefreshTTL       time.Duration
	ProviderBaseURL  string
	FrontendOrigin   string
	MigrationsDir    string
}

func Load() (Config, error) {
	accessTTL, err := getDuration("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, fmt.Errorf("parse ACCESS_TOKEN_TTL: %w", err)
	}

	refreshTTL, err := getDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("parse REFRESH_TOKEN_TTL: %w", err)
	}

	cfg := Config{
		AppName:          getEnv("APP_NAME", "crypto-parser"),
		Environment:      getEnv("APP_ENV", "development"),
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/crypto_parser?sslmode=disable"),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "local-access-secret-change-me"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "local-refresh-secret-change-me"),
		AccessTTL:        accessTTL,
		RefreshTTL:       refreshTTL,
		ProviderBaseURL:  getEnv("RATES_PROVIDER_URL", "https://api.coingecko.com/api/v3"),
		FrontendOrigin:   getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		MigrationsDir:    getEnv("MIGRATIONS_DIR", "./migrations"),
	}

	if cfg.JWTAccessSecret == "" || cfg.JWTRefreshSecret == "" {
		return Config{}, fmt.Errorf("JWT secrets must not be empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	return time.ParseDuration(value)
}
