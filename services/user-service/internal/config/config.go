package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name string `yaml:"Name"`
	Host string `yaml:"Host"`
	Port int    `yaml:"Port"`
	Auth struct {
		AccessSecret string `yaml:"AccessSecret"`
		AccessExpire int64  `yaml:"AccessExpire"`
	} `yaml:"Auth"`
	Database struct {
		DataSource string `yaml:"DataSource"`
	} `yaml:"Database"`
	VideoDir          string `yaml:"VideoDir"`
	InternalToken     string
	Version           string
	Commit            string
	BuildTime         string
	LogLevel          string
	RequestTimeout    time.Duration
	ContentServiceURL string `yaml:"ContentServiceURL"`
}

func Load(path string) (Config, error) {
	var c Config
	data, err := os.ReadFile(path)
	if err == nil {
		err = yaml.Unmarshal(data, &c)
	}
	if err != nil && !os.IsNotExist(err) {
		return c, err
	}
	c.Name = env("SERVICE_NAME", fallback(c.Name, "user-service"))
	c.Host = fallback(c.Host, "0.0.0.0")
	if p := os.Getenv("PORT"); p != "" {
		c.Port, _ = strconv.Atoi(p)
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	c.Database.DataSource = env("DATABASE_DSN", c.Database.DataSource)
	c.Auth.AccessSecret = env("JWT_SECRET", c.Auth.AccessSecret)
	c.InternalToken = os.Getenv("INTERNAL_API_TOKEN")
	c.Version = env("SERVICE_VERSION", "microservice-0.1.0")
	c.Commit = env("COMMIT_SHA", "development")
	c.BuildTime = env("BUILD_TIME", time.Now().Format(time.RFC3339))
	c.LogLevel = env("LOG_LEVEL", "info")
	c.ContentServiceURL = env("CONTENT_SERVICE_URL", c.ContentServiceURL)
	c.RequestTimeout = 2 * time.Second
	if value := os.Getenv("REQUEST_TIMEOUT"); value != "" {
		parsed, parseErr := time.ParseDuration(value)
		if parseErr != nil || parsed <= 0 || parsed > 2*time.Second {
			return c, fmt.Errorf("REQUEST_TIMEOUT must be a duration between 0 and 2s")
		}
		c.RequestTimeout = parsed
	}
	if c.Database.DataSource == "" || c.Auth.AccessSecret == "" || c.InternalToken == "" || c.ContentServiceURL == "" {
		return c, fmt.Errorf("DATABASE_DSN, JWT_SECRET, INTERNAL_API_TOKEN and CONTENT_SERVICE_URL are required")
	}
	return c, nil
}

func env(key, value string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return value
}
func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
