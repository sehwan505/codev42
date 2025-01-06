package storage

import (
	"context"
	"fmt"
	"log"

	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
)

// MilvusConnection : Milvus 연동 구조체
type MilvusConnection struct {
	Client milvusClient.Client
}

// NewMilvusConnection : Milvus 연결 생성
func NewMilvusConnection(ctx context.Context, milvusAddr string) (*MilvusConnection, error) {
	client, err := milvusClient.NewGrpcClient(ctx, milvusAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to milvus: %w", err)
	}
	return &MilvusConnection{Client: client}, nil
}

// Close : Milvus 연결 해제
func (m *MilvusConnection) Close() error {
	if m.Client != nil {
		err := m.Client.Close()
		if err != nil {
			log.Printf("failed to close milvus client: %v", err)
			return err
		}
	}
	return nil
}
