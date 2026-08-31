package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name     string `yaml:"Name"`
	Host     string `yaml:"Host"`
	Port     int    `yaml:"Port"`
	Database struct {
		DataSource string `yaml:"DataSource"`
	} `yaml:"Database"`
	Auth struct {
		AccessSecret string `yaml:"AccessSecret"`
	} `yaml:"Auth"`
	Dependencies struct {
		UserBaseURL    string        `yaml:"UserBaseURL"`
		ContentBaseURL string        `yaml:"ContentBaseURL"`
		InternalToken  string        `yaml:"InternalToken"`
		Timeout        time.Duration `yaml:"-"`
		TimeoutMS      int           `yaml:"TimeoutMS"`
	} `yaml:"Dependencies"`
	Live struct {
		RTMPHost string `yaml:"RTMPHost"`
	} `yaml:"Live"`
	Build struct {
		Version string `yaml:"Version"`
		GitSHA  string `yaml:"GitSHA"`
		Time    string `yaml:"Time"`
	} `yaml:"Build"`
}

func Load(path string) (Config, error) {
	c := Config{Name: "engagement-service", Host: "0.0.0.0", Port: 8080}
	c.Dependencies.TimeoutMS = 1500
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return c, fmt.Errorf("read config: %w", err)
		}
		if err := yaml.Unmarshal(data, &c); err != nil {
			return c, fmt.Errorf("parse config: %w", err)
		}
	}
	setString(&c.Name, "SERVICE_NAME")
	setString(&c.Database.DataSource, "DATABASE_DSN")
	setString(&c.Auth.AccessSecret, "JWT_SECRET")
	setString(&c.Dependencies.UserBaseURL, "USER_SERVICE_URL")
	setString(&c.Dependencies.ContentBaseURL, "CONTENT_SERVICE_URL")
	setString(&c.Dependencies.InternalToken, "INTERNAL_API_TOKEN")
	setString(&c.Live.RTMPHost, "SRS_RTMP_HOST")
	setString(&c.Build.Version, "SERVICE_VERSION")
	setString(&c.Build.GitSHA, "COMMIT_SHA")
	setString(&c.Build.Time, "BUILD_TIME")
	if value := os.Getenv("PORT"); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return c, fmt.Errorf("invalid PORT: %w", err)
		}
		c.Port = port
	}
	if value := os.Getenv("REQUEST_TIMEOUT"); value != "" {
		d, err := time.ParseDuration(value)
		if err != nil || d <= 0 || d > 2*time.Second {
			return c, fmt.Errorf("REQUEST_TIMEOUT must be in (0,2s]")
		}
		c.Dependencies.Timeout = d
		c.Dependencies.TimeoutMS = int(d / time.Millisecond)
	}
	if c.Database.DataSource == "" {
		return c, fmt.Errorf("DATABASE_DSN is required")
	}
	if c.Auth.AccessSecret == "" {
		return c, fmt.Errorf("JWT_SECRET is required")
	}
	if c.Dependencies.InternalToken == "" {
		return c, fmt.Errorf("INTERNAL_API_TOKEN is required")
	}
	if c.Dependencies.TimeoutMS < 1 || c.Dependencies.TimeoutMS > 2000 {
		return c, fmt.Errorf("Dependencies.TimeoutMS must be 1..2000")
	}
	if c.Dependencies.Timeout == 0 {
		c.Dependencies.Timeout = time.Duration(c.Dependencies.TimeoutMS) * time.Millisecond
	}
	return c, nil
}

func setString(target *string, key string) {
	if value := os.Getenv(key); value != "" {
		*target = value
	}
}
