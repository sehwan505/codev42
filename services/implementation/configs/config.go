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

	GRPCPort string

	// 서비스 간 통신 엔드포인트
	PlanServiceAddr     string
	DiagramServiceAddr  string
	AnalyzerServiceAddr string
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

		MySQLUser:     GetEnv("MYSQL_USER", "mainuser"),
		MySQLPassword: GetEnv("MYSQL_PASSWORD", "user123"),
		MySQLHost:     GetEnv("MYSQL_HOST", "localhost"),
		MySQLPort:     GetEnv("MYSQL_PORT", "3306"),
		MySQLDB:       GetEnv("MYSQL_DB", "codev"),

		GRPCPort: GetEnv("GRPC_PORT", "9092"),

		// 서비스 엔드포인트
		PlanServiceAddr:     GetEnv("PLAN_SERVICE_ADDR", "localhost:9091"),
		DiagramServiceAddr:  GetEnv("DIAGRAM_SERVICE_ADDR", "localhost:9093"),
		AnalyzerServiceAddr: GetEnv("ANALYZER_SERVICE_ADDR", "localhost:9094"),
	}

	if config.OpenAiKey == "" {
		return nil, fmt.Errorf("environment variable OPENAI_API_KEY is required but not set")
	}

	if config.MySQLPassword == "" {
		return nil, fmt.Errorf("environment variable MYSQL_PASSWORD is required but not set")
	}

	return config, nil
}
