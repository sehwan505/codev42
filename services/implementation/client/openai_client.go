package client

import (
	"net/http"
	"sync"
	"time"

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
		// 커스텀 HTTP 클라이언트 생성
		httpClient := &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     20,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		}

		instance = &OpenAIClient{
			client: openai.NewClient(
				option.WithAPIKey(apiKey),
				option.WithHTTPClient(httpClient), // 커스텀 HTTP 클라이언트 사용
			),
		}
	})
	return instance
}

func (o *OpenAIClient) Client() *openai.Client {
	return o.client
}
