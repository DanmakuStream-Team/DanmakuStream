package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func validEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_DSN", "user:pass@tcp(mysql:3306)/user_db")
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("INTERNAL_API_TOKEN", "internal-token")
	t.Setenv("CONTENT_SERVICE_URL", "http://content-service:8080")
}

func TestLoadDay7Dependencies(t *testing.T) {
	validEnvironment(t)
	t.Setenv("REQUEST_TIMEOUT", "1500ms")
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ContentServiceURL != "http://content-service:8080" || cfg.RequestTimeout != 1500*time.Millisecond {
		t.Fatalf("config=%+v", cfg)
	}
}

func TestLoadRequiresContentServiceURL(t *testing.T) {
	validEnvironment(t)
	if err := os.Unsetenv("CONTENT_SERVICE_URL"); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("expected missing CONTENT_SERVICE_URL error")
	}
}

func TestLoadRejectsExcessiveTimeout(t *testing.T) {
	validEnvironment(t)
	t.Setenv("REQUEST_TIMEOUT", "3s")
	if _, err := Load(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("expected REQUEST_TIMEOUT validation error")
	}
}
