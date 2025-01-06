package storage

import (
	"context"

	"github.com/pinecone-io/go-pinecone/pinecone"
)

// PineconeConnection holds your Pinecone client and any other config needed.
type PineconeConnection struct {
	Client *pinecone.Client
}

// NewPineconeConnection initializes a connection to Pinecone.
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

// Close closes the Pinecone client connection, if needed.
func (p *PineconeConnection) Close() error {
	// p.client.Close() // if Pinecone supports a Close()
	return nil
}
