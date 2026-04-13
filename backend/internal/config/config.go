package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	JWTTTL              time.Duration
	BcryptCost          int
	AppEnv              string
	DBMaxOpenConns      int
	DBMaxIdleConns      int
	DBConnMaxLifetime   time.Duration
	HTTPReadTimeout     time.Duration
	HTTPWriteTimeout    time.Duration
	HTTPIdleTimeout     time.Duration
	MaxRequestBodyBytes int64
}

func Load() (Config, error) {
	cfg := Config{
		Port:        envOrDefault("BACKEND_PORT", "8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:   strings.TrimSpace(os.Getenv("JWT_SECRET")),
		AppEnv:      envOrDefault("APP_ENV", "development"),
	}
	isProduction := strings.EqualFold(cfg.AppEnv, "production")

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	ttlHours, err := intEnvSetting("JWT_TTL_HOURS", 24, isProduction)
	if err != nil {
		return Config{}, err
	}
	cfg.JWTTTL = time.Duration(ttlHours) * time.Hour

	cfg.BcryptCost, err = intEnvSetting("BCRYPT_COST", 12, isProduction)
	if err != nil {
		return Config{}, err
	}
	if cfg.BcryptCost < 10 || cfg.BcryptCost > 14 {
		return Config{}, fmt.Errorf("BCRYPT_COST must be between 10 and 14")
	}

	cfg.DBMaxOpenConns, err = intEnvSetting("DB_MAX_OPEN_CONNS", 10, isProduction)
	if err != nil {
		return Config{}, err
	}
	cfg.DBMaxIdleConns, err = intEnvSetting("DB_MAX_IDLE_CONNS", 5, isProduction)
	if err != nil {
		return Config{}, err
	}
	if cfg.DBMaxIdleConns > cfg.DBMaxOpenConns {
		return Config{}, fmt.Errorf("DB_MAX_IDLE_CONNS cannot be greater than DB_MAX_OPEN_CONNS")
	}

	cfg.DBConnMaxLifetime, err = durationEnvSetting("DB_CONN_MAX_LIFETIME", 30*time.Minute, isProduction)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTPReadTimeout, err = durationEnvSetting("HTTP_READ_TIMEOUT", 10*time.Second, isProduction)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTPWriteTimeout, err = durationEnvSetting("HTTP_WRITE_TIMEOUT", 15*time.Second, isProduction)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTPIdleTimeout, err = durationEnvSetting("HTTP_IDLE_TIMEOUT", 60*time.Second, isProduction)
	if err != nil {
		return Config{}, err
	}

	maxBodyBytes, err := intEnvSetting("HTTP_MAX_REQUEST_BODY_BYTES", 1<<20, isProduction)
	if err != nil {
		return Config{}, err
	}
	cfg.MaxRequestBodyBytes = int64(maxBodyBytes)

	return cfg, nil
}

func (c Config) LogLevel() slog.Level {
	if strings.EqualFold(c.AppEnv, "production") {
		return slog.LevelInfo
	}
	return slog.LevelDebug
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func intEnvSetting(key string, fallback int, requireExplicit bool) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		if requireExplicit {
			return 0, fmt.Errorf("%s must be set when APP_ENV=production", key)
		}
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", key)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return parsed, nil
}

func durationEnvSetting(key string, fallback time.Duration, requireExplicit bool) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		if requireExplicit {
			return 0, fmt.Errorf("%s must be set when APP_ENV=production", key)
		}
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration", key)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return parsed, nil
}
