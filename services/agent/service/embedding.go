package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

type embeddingAgent struct {
	Client *openai.Client
}

func NewEmbeddingAgent(apiKey string) *embeddingAgent {
	openaiClient := GetClient(apiKey)

	return &embeddingAgent{
		Client: openaiClient.Client(),
	}
}

func (agent embeddingAgent) GetEmbedding(plans []string) ([][]float32, error) {

	var wg sync.WaitGroup
	resultChan := make(chan []float32, len(plans))
	errorChan := make(chan error, len(plans))
	for _, annotation := range plans {
		wg.Add(1)
		go func(annotation string) {
			defer wg.Done()
			response, err := agent.Client.Embeddings.New(context.TODO(), openai.EmbeddingNewParams{
				Input:          openai.F[openai.EmbeddingNewParamsInputUnion](shared.UnionString(annotation)),
				Model:          openai.F(openai.EmbeddingModelTextEmbedding3Small),
				Dimensions:     openai.F(int64(128)),
				EncodingFormat: openai.F(openai.EmbeddingNewParamsEncodingFormatFloat),
				User:           openai.F("hado_coder"),
			})
			fmt.Print(response)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- nil
		}(annotation)
	}

	return nil, nil
}
