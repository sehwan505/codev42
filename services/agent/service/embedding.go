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

func float64ToFloat32(input []float64) []float32 {
	output := make([]float32, len(input))
	for i, v := range input {
		output[i] = float32(v)
	}
	return output
}

func (agent embeddingAgent) GenerateEmbedding(codes map[int64]string) (map[int64][]float32, error) {
	type embeddingResult struct {
		ID        int64
		Embedding []float32
	}
	var wg sync.WaitGroup

	resultChan := make(chan embeddingResult, len(codes))
	errorChan := make(chan error, len(codes))
	for id, chunk := range codes {
		wg.Add(1)
		go func(chunk string, id int64) {
			defer wg.Done()
			response, err := agent.Client.Embeddings.New(context.TODO(), openai.EmbeddingNewParams{
				Input:          openai.F[openai.EmbeddingNewParamsInputUnion](shared.UnionString(chunk)),
				Model:          openai.F(openai.EmbeddingModelTextEmbedding3Small),
				Dimensions:     openai.F(int64(128)),
				EncodingFormat: openai.F(openai.EmbeddingNewParamsEncodingFormatFloat),
				User:           openai.F("hado_coder"),
			})
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- embeddingResult{
				ID:        id,
				Embedding: float64ToFloat32(response.Data[0].Embedding),
			}
		}(chunk, id)
	}

	wg.Wait()
	close(resultChan)
	close(errorChan)
	results := make(map[int64][]float32, len(resultChan))
	for result := range resultChan {
		results[result.ID] = result.Embedding
	}
	if len(errorChan) > 0 {
		var errors []string
		for err := range errorChan {
			errors = append(errors, err.Error())
		}
		return nil, fmt.Errorf("failed to implement plan: %v", errors)
	}
	return results, nil
}
