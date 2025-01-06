package configs

import (
	"fmt"
	"os"
)

type Config struct {
	OpenAiKey string

	MySQLUser     string
	MySQLPassword string
	MySQLHost     string
	MySQLPort     string
	MySQLDB       string

	PineconeApiKey string

	// Milvus
	MilvusHost string
	MilvusPort string

	// gRPC
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

		MySQLUser:     GetEnv("MYSQL_USER", "root"),
		MySQLPassword: GetEnv("MYSQL_PASSWORD", ""),
		MySQLHost:     GetEnv("MYSQL_HOST", "localhost"),
		MySQLPort:     GetEnv("MYSQL_PORT", "3306"),
		MySQLDB:       GetEnv("MYSQL_DB", "test"),

		PineconeApiKey: GetEnv("PINECONE_API_KEY", ""),

		MilvusHost: GetEnv("MILVUS_HOST", "localhost"),
		MilvusPort: GetEnv("MILVUS_PORT", "19530"),

		GRPCPort: GetEnv("GRPC_PORT", "9090"),
	}

	if config.OpenAiKey == "" || config.MySQLPassword == "" || config.PineconeApiKey == "" {
		return nil, fmt.Errorf("environment variable OPENAI_API_KEY is required but not set")
	}

	return config, nil
}
