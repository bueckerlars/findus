package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port         int
	DataDir      string
	BaseURL      string
	JWTSecret    string
	CookieSecure bool
	LogLevel     string
	LogFormat    string // text (default), json
}

func Load() (Config, error) {
	c := Config{
		Port:         8080,
		DataDir:      "./data",
		BaseURL:      "http://localhost:8080",
		CookieSecure: false,
		LogLevel:     "info",
		LogFormat:    "text",
	}
	if v := os.Getenv("FINDUS_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("FINDUS_PORT: %w", err)
		}
		c.Port = p
	}
	if v := strings.TrimSpace(os.Getenv("FINDUS_DATA_DIR")); v != "" {
		c.DataDir = v
	}
	if v := strings.TrimSpace(os.Getenv("FINDUS_BASE_URL")); v != "" {
		c.BaseURL = strings.TrimRight(v, "/")
	}
	if v := strings.TrimSpace(os.Getenv("FINDUS_JWT_SECRET")); v != "" {
		c.JWTSecret = v
	}
	if v := strings.TrimSpace(os.Getenv("FINDUS_COOKIE_SECURE")); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return Config{}, fmt.Errorf("FINDUS_COOKIE_SECURE: %w", err)
		}
		c.CookieSecure = b
	}
	if v := strings.TrimSpace(os.Getenv("FINDUS_LOG_LEVEL")); v != "" {
		c.LogLevel = v
	}
	if v := strings.TrimSpace(os.Getenv("FINDUS_LOG_FORMAT")); v != "" {
		c.LogFormat = v
	}
	return c, nil
}
