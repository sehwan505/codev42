package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var BasePath string

type Config struct {
	OpenAiKey string `json:"OPENAI_API_KEY"`

	MySQLUser     string `json:"MYSQL_USER"`
	MySQLPassword string `json:"MYSQL_PASSWORD"`
	MySQLHost     string `json:"MYSQL_HOST"`
	MySQLPort     string `json:"MYSQL_PORT"`
	MySQLDB       string `json:"MYSQL_DB"`

	// Milvus
	MilvusHost string `json:"MILVUS_HOST"`
	MilvusPort string `json:"MILVUS_PORT"`

	// gRPC
	GRPCPort string `json:"GRPC_PORT"`
}

func GetConfig() (*Config, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get runtime caller information")
	}

	configFilePath := filepath.Join(filepath.Dir(filename), "config.json")
	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil

}
