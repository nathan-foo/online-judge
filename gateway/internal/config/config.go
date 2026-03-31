package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

const MaxUploadSize = 1 << 20

type Config struct {
	Auth           AuthConfig
	Redis          RedisConfig
	AllowedOrigins []string
	Routes         []RouteConfig
}

type AuthConfig struct {
	ClerkSecretKey string
}

type RedisConfig struct {
	RedisUrl        string
	RateLimitGlobal int
}

type RouteConfig struct {
	Prefix        string
	ServiceUrl    string
	RequireAuth   bool
	RateLimit     int
	MaxUploadSize int64
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Auth: AuthConfig{
			getConfig("CLERK_SECRET_KEY"),
		},
		Redis: RedisConfig{
			getConfig("REDIS_URL"),
			getConfigInt("RATE_LIMIT_GLOBAL", 100),
		},
		AllowedOrigins: parseOrigins(getConfig("ALLOWED_ORIGINS")),
		Routes:         Routes,
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
	if cfg.Redis.RedisUrl == "" {
		return errors.New("REDIS_URL is required")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return errors.New("ALLOWED_ORIGINS is required")
	}
	for _, route := range cfg.Routes {
		if route.Prefix == "" {
			return errors.New("Route prefix is required")
		}
		if route.ServiceUrl == "" {
			return errors.New("Service URL is required")
		}
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
