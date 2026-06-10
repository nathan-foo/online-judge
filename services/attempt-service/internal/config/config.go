package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Database DatabaseConfig
	Port     int
}

type DatabaseConfig struct {
	DatabaseUrl string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			DatabaseUrl: databaseUrl(),
		},
		Port: getConfigInt("PORT", 8000),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	if cfg.Database.DatabaseUrl == "" {
		return errors.New("DATABASE_URL or POSTGRES_PASSWORD is required")
	}
	return nil
}

func databaseUrl() string {
	if url := getConfig("DATABASE_URL"); url != "" {
		return url
	}

	password := getConfig("POSTGRES_PASSWORD")
	if password == "" {
		return ""
	}

	return "postgres://postgres:" + password + "@postgres:5432/attempts"
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

func getConfigInt(key string, fallback int) int {
	val := getConfig(key)
	if val == "" {
		return fallback
	}

	num, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return num
}
