package repo

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"codev42/services/agent/storage"

	pinecone "github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

// PineconeRepo implements the VectorDB interface for Pinecone.
type PineconeRepo struct {
	*storage.PineconeConnection
	idxConnection *pinecone.IndexConnection
}

func NewPineconeRepo(conn *storage.PineconeConnection) *PineconeRepo {
	return &PineconeRepo{conn, nil}
}

func (r *PineconeRepo) InitCollection(ctx context.Context, collectionName string, vectorDim int) error {
	desc, err := r.Client.DescribeIndex(ctx, collectionName)
	if err == nil && desc != nil {
		log.Printf("Pinecone index '%s' already exists. Dimension: %d", collectionName, desc.Dimension)
		idx, err := r.Client.DescribeIndex(ctx, collectionName)
		if err != nil {
			log.Fatalf("Failed to describe index \"%v\": %v", idx.Name, err)
		}

		idxConnection, err := r.Client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
		if err != nil {
			log.Fatalf("Failed to create IndexConnection for Host: %v: %v", idx.Host, err)
		}
		r.idxConnection = idxConnection
		return nil
	}
	// podIndexMetadata := &pinecone.PodSpecMetadataConfig{
	// 	Indexed: &[]string{"id"},
	// }
	// createReq := pinecone.CreatePodIndexRequest{
	// 	Name:           collectionName,
	// 	Dimension:      int32(vectorDim),
	// 	Metric:         pinecone.Cosine,
	// 	Environment:    "us-west1-gcp",
	// 	PodType:        "s1",
	// 	MetadataConfig: podIndexMetadata,
	// }
	// idx, err := r.Client.CreatePodIndex(ctx, &createReq)
	idx, err := r.Client.CreateServerlessIndex(ctx, &pinecone.CreateServerlessIndexRequest{
		Name:      collectionName,
		Dimension: 128,
		Metric:    pinecone.Cosine,
		Cloud:     pinecone.Aws,
		Region:    "us-east-1",
	})

	if err != nil {
		return fmt.Errorf("failed to create Pinecone index: %w", err)
	}
	idxConnection, err := r.Client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v: %v", idx.Host, err)
	}
	r.idxConnection = idxConnection
	log.Printf("Created Pinecone index '%s' with dimension %d", collectionName, vectorDim)
	return nil
}

func (r *PineconeRepo) InsertEmbedding(ctx context.Context, collectionName string, id string, embedding []float32) error {
	metadataMap := map[string]interface{}{
		"genre": "classical",
	}
	metadata, err := structpb.NewStruct(metadataMap)
	if err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	vectors := []*pinecone.Vector{
		{
			Id:       id,
			Values:   embedding,
			Metadata: metadata,
		},
	}

	if _, err := r.idxConnection.UpsertVectors(ctx, vectors); err != nil {
		return fmt.Errorf("failed to insert vectors: %w", err)
	}
	return nil
}

func (r *PineconeRepo) DeleteByID(ctx context.Context, collectionName string, id string) error {
	if err := r.idxConnection.DeleteVectorsById(ctx, []string{id}); err != nil {
		return fmt.Errorf("벡터 삭제 실패: %w", err)
	}
	return nil
}

func (r *PineconeRepo) SearchByVector(ctx context.Context, collectionName string, searchVector []float32, topK int) ([]int64, error) {

	metadataMap := map[string]interface{}{
		"genre": map[string]interface{}{
			"$eq": "documentary",
		},
		"year": 2019,
	}

	metadataFilter, err := structpb.NewStruct(metadataMap)
	if err != nil {
		log.Fatalf("Failed to create metadataFilter: %v", err)
	}

	res, err := r.idxConnection.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:         searchVector,
		TopK:           3,
		MetadataFilter: metadataFilter,
		IncludeValues:  true,
	})
	if err != nil {
		log.Fatalf("Error encountered when querying by vector: %v", err)
	}
	var ids []int64
	for _, match := range res.Matches {
		id, err := strconv.ParseInt(match.Vector.Id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ID: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Close cleans up resources associated x the Pinecone client.
func (r *PineconeRepo) Close() error {
	log.Println("PineconeRepo.Close() called - no resources to close.")
	return nil
}
