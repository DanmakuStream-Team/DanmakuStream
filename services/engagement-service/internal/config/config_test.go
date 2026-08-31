package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesStandardEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("Name: ignored\nHost: 0.0.0.0\nPort: 8080\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SERVICE_NAME", "engagement-service")
	t.Setenv("DATABASE_DSN", "user:pass@tcp(mysql:3306)/engagement_db")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("INTERNAL_API_TOKEN", "internal-test-token")
	t.Setenv("COMMIT_SHA", "abcdef0")
	t.Setenv("REQUEST_TIMEOUT", "1500ms")
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "engagement-service" || c.Build.GitSHA != "abcdef0" {
		t.Fatalf("unexpected config: %+v", c)
	}
	if c.Dependencies.TimeoutMS != 1500 {
		t.Fatalf("timeout=%d", c.Dependencies.TimeoutMS)
	}
}

func TestLoadRejectsTimeoutOverTwoSeconds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	_ = os.WriteFile(path, []byte("Host: 0.0.0.0\n"), 0600)
	t.Setenv("DATABASE_DSN", "dsn")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("INTERNAL_API_TOKEN", "token")
	t.Setenv("REQUEST_TIMEOUT", "3s")
	if _, err := Load(path); err == nil {
		t.Fatal("expected timeout validation error")
	}
}
