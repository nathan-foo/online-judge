package config

import (
	"errors"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Auth      AuthConfig
	Endpoints EndpointConfig
}

type AuthConfig struct {
	CLERK_SECRET_KEY string
}

type EndpointConfig struct {
	TEST_SERVICE_URL string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Auth: AuthConfig{
			getConfig("CLERK_SECRET_KEY"),
		},
		Endpoints: EndpointConfig{
			getConfig("TEST_SERVICE_URL"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	if cfg.Auth.CLERK_SECRET_KEY == "" {
		return errors.New("CLERK_SECRET_KEY is required")
	}
	if cfg.Endpoints.TEST_SERVICE_URL == "" {
		return errors.New("TEST_SERVICE_URL is required")
	}
	return nil
}

func getConfig(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	path := "/run/secrets/" + strings.ToLower(key)
	data, err := os.ReadFile(path)

	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}
