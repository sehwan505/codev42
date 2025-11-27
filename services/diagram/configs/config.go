package configs

import (
	"fmt"
	"os"
)

type Config struct {
	OpenAiKey string
	GRPCPort  string
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetConfig() (*Config, error) {
	config := &Config{
		OpenAiKey: GetEnv("OPENAI_API_KEY", ""),
		GRPCPort:  GetEnv("GRPC_PORT", "9093"),
	}

	if config.OpenAiKey == "" {
		return nil, fmt.Errorf("environment variable OPENAI_API_KEY is required but not set")
	}

	return config, nil
}
