package configs

import (
	"fmt"
	"os"
)

type Config struct {
	OpenAiKey string

	// Plan Service gRPC endpoint
	PlanServiceAddr string

	// Redis for Job Queue
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	GRPCPort string
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

		PlanServiceAddr: GetEnv("PLAN_SERVICE_ADDR", "localhost:9091"),

		RedisAddr:     GetEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: GetEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,

		GRPCPort: GetEnv("GRPC_PORT", "9092"),
	}

	if config.OpenAiKey == "" {
		return nil, fmt.Errorf("environment variable OPENAI_API_KEY is required but not set")
	}

	return config, nil
}
