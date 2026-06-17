package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Port        int
	RabbitMQURL string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:        getConfigInt("PORT", 8000),
		RabbitMQURL: fmt.Sprintf("amqp://online_judge:%s@rabbitmq:5672/", getConfig("RABBITMQ_PASSWORD")),
	}

	return cfg, nil
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
