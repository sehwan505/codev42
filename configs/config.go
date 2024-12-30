package configs

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	GitToken  string `json:"GIT_TOKEN"`
	GitUserID string `json:"GIT_USERID"`
	GitRepo   string `json:"GIT_REPO"`

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

	// Gin HTTP
	HTTPPort string `json:"HTTP_PORT"`
}

func GetConfig() (*Config, error) {
	file, err := os.Open("configs/config.json")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return nil, err
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config:", err)
		return nil, err
	}
	return &config, nil

}
