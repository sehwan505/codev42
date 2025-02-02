package config

import "os"

type Config struct {
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
	return &Config{
		GRPCPort: GetEnv("GRPC_PORT", "9091"),
	}, nil
}
