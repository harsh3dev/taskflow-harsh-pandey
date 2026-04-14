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
	JWTActiveKeyID      string
	JWTSigningKeys      map[string]string
	JWTIssuer           string
	JWTAudience         string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	BcryptCost          int
	AppEnv              string
	DBMaxOpenConns      int
	DBMaxIdleConns      int
	DBConnMaxLifetime   time.Duration
	HTTPReadTimeout     time.Duration
	HTTPWriteTimeout    time.Duration
	HTTPIdleTimeout     time.Duration
	MaxRequestBodyBytes int64
	// Notifications
	NotificationsEnabled bool
	RedisURL             string
	SMTPHost             string
	SMTPPort             int
	SMTPUsername         string
	SMTPPassword         string
	SMTPFromAddress      string
}

func Load() (Config, error) {
	cfg := Config{
		Port:        envOrDefault("BACKEND_PORT", "8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AppEnv:      envOrDefault("APP_ENV", "development"),
		JWTIssuer:   envOrDefault("JWT_ISSUER", "taskflow"),
		JWTAudience: envOrDefault("JWT_AUDIENCE", "taskflow-api"),
	}
	isProduction := strings.EqualFold(cfg.AppEnv, "production")

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	signingKeys, activeKeyID, err := loadJWTSigningKeys()
	if err != nil {
		return Config{}, err
	}
	cfg.JWTSigningKeys = signingKeys
	cfg.JWTActiveKeyID = activeKeyID

	cfg.AccessTokenTTL, err = accessTokenTTLSetting(isProduction)
	if err != nil {
		return Config{}, err
	}
	cfg.RefreshTokenTTL, err = durationEnvSetting("REFRESH_TOKEN_TTL", 30*24*time.Hour, isProduction)
	if err != nil {
		return Config{}, err
	}

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

	// Notifications
	cfg.NotificationsEnabled = strings.EqualFold(envOrDefault("NOTIFICATIONS_ENABLED", "false"), "true")
	cfg.RedisURL = envOrDefault("REDIS_URL", "redis://localhost:6379")
	cfg.SMTPHost = envOrDefault("SMTP_HOST", "localhost")
	cfg.SMTPPort, err = intEnvSetting("SMTP_PORT", 1025, false)
	if err != nil {
		return Config{}, err
	}
	cfg.SMTPUsername = os.Getenv("SMTP_USERNAME")
	cfg.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	cfg.SMTPFromAddress = envOrDefault("SMTP_FROM", "noreply@taskflow.app")

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

func accessTokenTTLSetting(requireExplicit bool) (time.Duration, error) {
	if value := strings.TrimSpace(os.Getenv("ACCESS_TOKEN_TTL")); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return 0, fmt.Errorf("ACCESS_TOKEN_TTL must be a valid duration")
		}
		if parsed <= 0 {
			return 0, fmt.Errorf("ACCESS_TOKEN_TTL must be greater than zero")
		}
		return parsed, nil
	}

	if legacyValue := strings.TrimSpace(os.Getenv("JWT_TTL_HOURS")); legacyValue != "" {
		parsed, err := strconv.Atoi(legacyValue)
		if err != nil {
			return 0, fmt.Errorf("JWT_TTL_HOURS must be a valid integer")
		}
		if parsed <= 0 {
			return 0, fmt.Errorf("JWT_TTL_HOURS must be greater than zero")
		}
		return time.Duration(parsed) * time.Hour, nil
	}

	if requireExplicit {
		return 0, fmt.Errorf("ACCESS_TOKEN_TTL must be set when APP_ENV=production")
	}
	return 15 * time.Minute, nil
}

func loadJWTSigningKeys() (map[string]string, string, error) {
	activeKeyID := strings.TrimSpace(os.Getenv("JWT_ACTIVE_KEY_ID"))
	encodedKeys := strings.TrimSpace(os.Getenv("JWT_SIGNING_KEYS"))
	legacySecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))

	keys := map[string]string{}
	if encodedKeys != "" {
		for _, entry := range strings.Split(encodedKeys, ",") {
			keyID, secret, found := strings.Cut(strings.TrimSpace(entry), ":")
			if !found || strings.TrimSpace(keyID) == "" || strings.TrimSpace(secret) == "" {
				return nil, "", fmt.Errorf("JWT_SIGNING_KEYS entries must be in kid:secret format")
			}
			if len(strings.TrimSpace(secret)) < 32 {
				return nil, "", fmt.Errorf("JWT signing secrets must be at least 32 characters")
			}
			keys[strings.TrimSpace(keyID)] = strings.TrimSpace(secret)
		}
	}

	if len(keys) == 0 {
		if legacySecret == "" {
			return nil, "", fmt.Errorf("JWT_SIGNING_KEYS or JWT_SECRET is required")
		}
		if len(legacySecret) < 32 {
			return nil, "", fmt.Errorf("JWT_SECRET must be at least 32 characters")
		}
		keys["default"] = legacySecret
		if activeKeyID == "" {
			activeKeyID = "default"
		}
	}

	if activeKeyID == "" {
		if _, ok := keys["default"]; ok {
			activeKeyID = "default"
		} else {
			return nil, "", fmt.Errorf("JWT_ACTIVE_KEY_ID is required when JWT_SIGNING_KEYS is set")
		}
	}

	if _, ok := keys[activeKeyID]; !ok {
		return nil, "", fmt.Errorf("JWT_ACTIVE_KEY_ID must reference a configured signing key")
	}

	return keys, activeKeyID, nil
}
