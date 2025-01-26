package handler

import (
	"context"
	"fmt"
	"strconv"

	"codev42/services/agent/configs"
	"codev42/services/agent/pb"
	"codev42/services/agent/service"
	"codev42/services/agent/storage"
)

type CodeHandler struct {
	pb.UnimplementedCodeServiceServer
	Config        configs.Config
	VectorDB      VectorDB
	RdbConnection *storage.RDBConnection
}

type VectorDB interface {
	InitCollection(ctx context.Context, collectionName string, vectorDim int) error
	InsertEmbedding(ctx context.Context, collectionName string, id string, embedding []float32) error
	SearchByVector(ctx context.Context, collectionName string, searchVector []float32, topK int) ([]int64, error)
	Close() error
}

func (c *CodeHandler) SaveCode(ctx context.Context, request *pb.SaveCodeRequest) (*pb.SaveCodeResponse, error) {
	agent := service.NewEmbeddingAgent(c.Config.OpenAiKey)

	codes, err := service.SaveCode(request.FilePath, request.Code, c.RdbConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to save code: %v", err)
	}
	embeddings, err := agent.GetEmbedding(codes) // add plan after
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %v", err)
	}
	for id, embedding := range embeddings {
		err = c.VectorDB.InsertEmbedding(ctx, "code", strconv.FormatInt(id, 10), embedding)
		if err != nil {
			return nil, fmt.Errorf("failed to insert embedding into VectorDB: %v", err)
		}
	}
	return &pb.SaveCodeResponse{
		Status: "success",
	}, nil
}
