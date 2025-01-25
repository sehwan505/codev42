package service

import (
	"sync"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIClient struct {
	client *openai.Client
}

var instance *OpenAIClient
var once sync.Once

func GetClient(apiKey string) *OpenAIClient {
	once.Do(func() {
		instance = &OpenAIClient{
			client: openai.NewClient(
				option.WithAPIKey(apiKey),
			),
		}
	})
	return instance
}

func (o *OpenAIClient) Client() *openai.Client {
	return o.client
}
