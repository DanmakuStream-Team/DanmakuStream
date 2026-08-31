package config

import "testing"

func validEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_DSN", "content_user:test@tcp(localhost:3306)/content_db")
	t.Setenv("JWT_SECRET", "jwt-test")
	t.Setenv("INTERNAL_API_TOKEN", "internal-test")
	t.Setenv("SERVICE_NAME", "content-service")
	t.Setenv("SERVICE_VERSION", "microservice-test")
	t.Setenv("COMMIT_SHA", "abcdef0")
	t.Setenv("BUILD_TIME", "2026-08-31T10:00:00+08:00")
	t.Setenv("PORT", "8080")
	t.Setenv("REQUEST_TIMEOUT", "2s")
}

func TestLoadValidatesRequiredEnvironment(t *testing.T) {
	validEnvironment(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServiceName != "content-service" || cfg.RequestTimeout.String() != "2s" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestLoadRejectsInvalidBuildMetadataAndTimeout(t *testing.T) {
	validEnvironment(t)
	t.Setenv("BUILD_TIME", "not-a-time")
	if _, err := Load(); err == nil {
		t.Fatal("invalid BUILD_TIME accepted")
	}
	validEnvironment(t)
	t.Setenv("REQUEST_TIMEOUT", "3s")
	if _, err := Load(); err == nil {
		t.Fatal("timeout above 2s accepted")
	}
}

func TestLoadRejectsMissingSecrets(t *testing.T) {
	validEnvironment(t)
	t.Setenv("JWT_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("missing JWT_SECRET accepted")
	}
}
