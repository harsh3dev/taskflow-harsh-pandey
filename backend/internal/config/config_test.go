package config

import (
	"testing"
	"time"
)

func TestLoadRejectsWeakJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/taskflow?sslmode=disable")
	t.Setenv("JWT_SECRET", "too-short")

	_, err := Load()
	if err == nil {
		t.Fatal("expected weak JWT secret to fail validation")
	}
}

func TestLoadRejectsInvalidIntegers(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/taskflow?sslmode=disable")
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("DB_MAX_OPEN_CONNS", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("expected invalid integer env var to fail validation")
	}
}

func TestLoadParsesProductionSettings(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/taskflow?sslmode=disable")
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("JWT_TTL_HOURS", "48")
	t.Setenv("BCRYPT_COST", "13")
	t.Setenv("DB_MAX_OPEN_CONNS", "20")
	t.Setenv("DB_MAX_IDLE_CONNS", "10")
	t.Setenv("DB_CONN_MAX_LIFETIME", "45m")
	t.Setenv("HTTP_READ_TIMEOUT", "12s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "18s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "75s")
	t.Setenv("HTTP_MAX_REQUEST_BODY_BYTES", "2048")
	t.Setenv("APP_ENV", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.JWTTTL != 48*time.Hour {
		t.Fatalf("unexpected JWT TTL: %v", cfg.JWTTTL)
	}
	if cfg.BcryptCost != 13 {
		t.Fatalf("unexpected bcrypt cost: %d", cfg.BcryptCost)
	}
	if cfg.DBMaxOpenConns != 20 || cfg.DBMaxIdleConns != 10 {
		t.Fatalf("unexpected DB pool config: open=%d idle=%d", cfg.DBMaxOpenConns, cfg.DBMaxIdleConns)
	}
	if cfg.DBConnMaxLifetime != 45*time.Minute {
		t.Fatalf("unexpected DB conn max lifetime: %v", cfg.DBConnMaxLifetime)
	}
	if cfg.HTTPReadTimeout != 12*time.Second || cfg.HTTPWriteTimeout != 18*time.Second || cfg.HTTPIdleTimeout != 75*time.Second {
		t.Fatalf("unexpected HTTP timeout config: read=%v write=%v idle=%v", cfg.HTTPReadTimeout, cfg.HTTPWriteTimeout, cfg.HTTPIdleTimeout)
	}
	if cfg.MaxRequestBodyBytes != 2048 {
		t.Fatalf("unexpected max request body bytes: %d", cfg.MaxRequestBodyBytes)
	}
}

func TestLoadRequiresExplicitRuntimeSettingsInProduction(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/taskflow?sslmode=disable")
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("APP_ENV", "production")

	_, err := Load()
	if err == nil {
		t.Fatal("expected production config without explicit runtime settings to fail")
	}
}
