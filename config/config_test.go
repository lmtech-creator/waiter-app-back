package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear env vars to test defaults
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.DSN == "" {
		t.Error("expected non-empty DSN")
	}
	// Should contain default values
	expected := "host=localhost port=5432 user=postgres password=postgres dbname=waiter sslmode=disable"
	if cfg.DSN != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, cfg.DSN)
	}
}

func TestLoad_CustomPort(t *testing.T) {
	os.Setenv("PORT", "3000")
	defer os.Unsetenv("PORT")

	cfg := Load()
	if cfg.Port != "3000" {
		t.Errorf("expected port 3000, got %s", cfg.Port)
	}
}

func TestLoad_DatabaseURL(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@host:5432/db")
	defer os.Unsetenv("DATABASE_URL")

	cfg := Load()
	if cfg.DSN != "postgres://user:pass@host:5432/db" {
		t.Errorf("expected DATABASE_URL value, got %s", cfg.DSN)
	}
}

func TestLoad_CustomDBEnvVars(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Setenv("DB_HOST", "myhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "myuser")
	os.Setenv("DB_PASSWORD", "mypass")
	os.Setenv("DB_NAME", "mydb")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
	}()

	cfg := Load()
	expected := "host=myhost port=5433 user=myuser password=mypass dbname=mydb sslmode=disable"
	if cfg.DSN != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, cfg.DSN)
	}
}

func TestGetEnv_WithValue(t *testing.T) {
	os.Setenv("TEST_KEY_WAITER", "hello")
	defer os.Unsetenv("TEST_KEY_WAITER")

	val := getEnv("TEST_KEY_WAITER", "default")
	if val != "hello" {
		t.Errorf("expected 'hello', got '%s'", val)
	}
}

func TestGetEnv_Fallback(t *testing.T) {
	os.Unsetenv("TEST_KEY_WAITER_MISSING")

	val := getEnv("TEST_KEY_WAITER_MISSING", "fallback")
	if val != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", val)
	}
}
