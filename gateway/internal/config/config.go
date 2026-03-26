package config

import (
	"errors"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	ClerkSecretKey string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		ClerkSecretKey: os.Getenv("CLERK_SECRET_KEY"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	if cfg.ClerkSecretKey == "" {
		return errors.New("CLERK_SECRET_KEY is required")
	}
	return nil
}
