package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type MarketDataMode string

const (
	MarketDataModeMock       MarketDataMode = "mock"
	MarketDataModeRepository MarketDataMode = "repository"
	MarketDataModeAkShare    MarketDataMode = "akshare"
)

type Config struct {
	Port           string
	DBHost         string
	DBPort         string
	DBName         string
	DBUser         string
	DBPassword     string
	CookieSecure   bool
	SessionTTL     time.Duration
	InitialCash    string
	MigrationsDir  string
	MarketDataMode MarketDataMode

	AkShareWebhookSecret string
	AkShareWebhookMaxAge time.Duration
	MarketDataStaleAfter time.Duration
	FrontendOrigin       string
}

func Load() (Config, error) {
	marketDataMode, err := parseMarketDataMode(env("MARKET_DATA_MODE", string(MarketDataModeMock)))
	if err != nil {
		return Config{}, err
	}
	webhookSecret := strings.TrimSpace(os.Getenv("AKSHARE_WEBHOOK_SECRET"))
	if marketDataMode == MarketDataModeAkShare && webhookSecret == "" {
		return Config{}, fmt.Errorf("AKSHARE_WEBHOOK_SECRET is required in akshare mode")
	}
	return Config{
		Port:           env("PORT", "8080"),
		DBHost:         env("DB_HOST", "mysql"),
		DBPort:         env("DB_PORT", "3306"),
		DBName:         env("DB_NAME", "ux_securities"),
		DBUser:         env("DB_USER", "ux_app"),
		DBPassword:     env("DB_PASSWORD", "local-dev-only"),
		CookieSecure:   envBool("COOKIE_SECURE", false),
		SessionTTL:     envDuration("SESSION_TTL", 24*time.Hour),
		InitialCash:    env("INITIAL_CASH_HKD", "1000000.0000"),
		MigrationsDir:  env("MIGRATIONS_DIR", "/app/db/migrations"),
		MarketDataMode: marketDataMode,

		AkShareWebhookSecret: webhookSecret,
		AkShareWebhookMaxAge: envDuration("AKSHARE_WEBHOOK_MAX_AGE", 30*time.Second),
		MarketDataStaleAfter: envDuration("MARKET_DATA_STALE_AFTER", 20*time.Minute),
		FrontendOrigin:       env("FRONTEND_ORIGIN", "http://localhost:3000"),
	}, nil
}

func (c Config) DatabaseDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
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

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseMarketDataMode(value string) (MarketDataMode, error) {
	mode := MarketDataMode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case MarketDataModeMock, MarketDataModeRepository, MarketDataModeAkShare:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported MARKET_DATA_MODE %q", value)
	}
}
