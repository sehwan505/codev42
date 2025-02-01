package handler

import (
	"context"
	"fmt"
	"strconv"

	"codev42/agent/configs"
	"codev42/agent/pb"
	"codev42/agent/service"
	"codev42/agent/storage"
)

type CodeHandler struct {
	pb.UnimplementedCodeServiceServer
	Config        configs.Config
	VectorDB      VectorDB
	RdbConnection *storage.RDBConnection
}

type VectorDB interface {
	InitCollection(ctx context.Context, collectionName string, vectorDim int32) error
	InsertEmbedding(ctx context.Context, collectionName string, id string, embedding []float32) error
	SearchByVector(ctx context.Context, collectionName string, searchVector []float32, topK int) ([]int64, error)
	DeleteByID(ctx context.Context, collectionName string, id string) error
	Close() error
}

func (c *CodeHandler) SaveCode(ctx context.Context, request *pb.SaveCodeRequest) (*pb.SaveCodeResponse, error) {
	agent := service.NewEmbeddingAgent(c.Config.OpenAiKey)

	saveCodeResult, err := service.SaveCode(request.Code, request.FilePath, c.RdbConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to save code: %v", err)
	}
	codes := make(map[int64]string)
	for id, result := range saveCodeResult {
		if result.IsNew || result.IsUpdated {
			codes[id] = result.Chunk
		}
		if result.IsUpdated {
			// 업데이트 된 경우 기존 코드 삭제
			c.VectorDB.DeleteByID(ctx, "code", strconv.FormatInt(id, 10))
		}
	}
	embeddings, err := agent.GenerateEmbedding(codes)
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
