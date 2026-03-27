package config

import (
	"errors"
	"os"

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
			os.Getenv("CLERK_SECRET_KEY"),
		},
		Endpoints: EndpointConfig{
			os.Getenv("TEST_SERVICE_URL"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	// if cfg.Auth.CLERK_SECRET_KEY == "" {
	// 	return errors.New("CLERK_SECRET_KEY is required")
	// }
	if cfg.Endpoints.TEST_SERVICE_URL == "" {
		return errors.New("TEST_SERVICE_URL is required")
	}
	return nil
}
