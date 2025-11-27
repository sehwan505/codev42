package storage

import (
	"context"

	"github.com/pinecone-io/go-pinecone/pinecone"
)

// PineconeConnection : Pinecone 클라이언트 및 기타 필요한 구성을 보관합니다.
type PineconeConnection struct {
	Client *pinecone.Client
}

func NewPineconeConnection(ctx context.Context, apiKey string) (*PineconeConnection, error) {
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	return &PineconeConnection{
		Client: pc,
	}, nil
}

func (p *PineconeConnection) Close() error {
	return nil
}
