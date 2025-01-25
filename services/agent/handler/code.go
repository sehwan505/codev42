package handler

import (
	"context"
	"fmt"

	"codev42/services/agent/configs"
	"codev42/services/agent/pb"
	"codev42/services/agent/service"
)

type CodeHandler struct {
	pb.UnimplementedCodeServiceServer
	Config   configs.Config
	VectorDB VectorDB
}

type VectorDB interface {
	InitCollection(ctx context.Context, collectionName string, vectorDim int) error
	InsertEmbedding(ctx context.Context, collectionName string, id string, embedding []float32) error
	SearchByVector(ctx context.Context, collectionName string, searchVector []float32, topK int) ([]int64, error)
	Close() error
}

func (c *CodeHandler) SaveCodeInVectordb(ctx context.Context, request *pb.SaveCodeRequest) (*pb.SaveCodeResponse, error) {
	// Create embeddings using the OpenAI agent
	agent := service.NewEmbeddingAgent(c.Config.OpenAiKey)
	embedding, err := agent.GetEmbedding(request.Plans)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %v", err)
	}

	// Print embedding for debugging
	fmt.Printf("Generated Embedding: %v\n", embedding)

	// Insert the embedding into VectorDB
	// err = c.VectorDB.InsertEmbedding(ctx, "code", request.Filename, embedding)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to insert embedding into VectorDB: %v", err)
	// }

	// Return success response
	return &pb.SaveCodeResponse{
		Status: "success",
	}, nil
}
