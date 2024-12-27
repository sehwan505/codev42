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
