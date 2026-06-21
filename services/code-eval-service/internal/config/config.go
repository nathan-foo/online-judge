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

	// Warm pool configuration
	Namespace        string
	ExecAgentImage   string
	ExecAgentTag     string
	PoolSize         int
	PoolSizes        map[string]int
	RuntimeClassName string
	Languages        []string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:        getConfigInt("PORT", 8000),
		RabbitMQURL: fmt.Sprintf("amqp://online_judge:%s@rabbitmq:5672/", getConfig("RABBITMQ_PASSWORD")),

		Namespace:        resolveNamespace(),
		ExecAgentImage:   getConfigDefault("EXEC_AGENT_IMAGE", "online-judge/exec-agent"),
		ExecAgentTag:     getConfigDefault("EXEC_AGENT_TAG", "latest"),
		PoolSize:         getConfigInt("POOL_SIZE", 2),
		PoolSizes:        getConfigIntMap("POOL_SIZES"),
		RuntimeClassName: getConfig("RUNTIME_CLASS_NAME"),
		Languages:        getConfigList("LANGUAGES", []string{"python", "c", "cpp", "java", "javascript", "go", "typescript", "kotlin", "rust", "csharp"}),
	}

	return cfg, nil
}

func resolveNamespace() string {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}
	const saNamespace = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	if data, err := os.ReadFile(saNamespace); err == nil {
		if ns := strings.TrimSpace(string(data)); ns != "" {
			return ns
		}
	}
	return "default"
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

func getConfigDefault(key, fallback string) string {
	if val := getConfig(key); val != "" {
		return val
	}
	return fallback
}

func getConfigList(key string, fallback []string) []string {
	val := getConfig(key)
	if val == "" {
		return fallback
	}

	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func getConfigIntMap(key string) map[string]int {
	out := map[string]int{}
	for part := range strings.SplitSeq(getConfig(key), ",") {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if k == "" || err != nil {
			continue
		}
		out[k] = n
	}
	return out
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
