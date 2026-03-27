package config

import (
	"errors"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Auth           AuthConfig
	Endpoints      EndpointConfig
	AllowedOrigins []string
}

type AuthConfig struct {
	ClerkSecretKey string
}

type EndpointConfig struct {
	TestServiceUrl string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Auth: AuthConfig{
			getConfig("CLERK_SECRET_KEY"),
		},
		Endpoints: EndpointConfig{
			getConfig("TEST_SERVICE_URL"),
		},
		AllowedOrigins: parseOrigins(getConfig("ALLOWED_ORIGINS")),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	if cfg.Auth.ClerkSecretKey == "" {
		return errors.New("CLERK_SECRET_KEY is required")
	}
	if cfg.Endpoints.TestServiceUrl == "" {
		return errors.New("TEST_SERVICE_URL is required")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return errors.New("ALLOWED_ORIGINS is required")
	}
	return nil
}

func parseOrigins(s string) []string {
	if s == "" {
		return nil
	}
	origins := strings.Split(s, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return origins
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
