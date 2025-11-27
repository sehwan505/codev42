package repo

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"codev42-plan/storage"

	pinecone "github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

// PineconeRepo는 Pinecone에 대한 VectorDB 인터페이스를 구현합니다.
type PineconeRepo struct {
	*storage.PineconeConnection
	idxConnection *pinecone.IndexConnection
}

func NewPineconeRepo(conn *storage.PineconeConnection) *PineconeRepo {
	return &PineconeRepo{conn, nil}
}

func (r *PineconeRepo) InitCollection(ctx context.Context, collectionName string, vectorDim int32) error {
	desc, err := r.Client.DescribeIndex(ctx, collectionName)
	if err == nil && desc != nil {
		log.Printf("Pinecone 인덱스 '%s'가 이미 존재합니다. 차원: %d", collectionName, desc.Dimension)
		idx, err := r.Client.DescribeIndex(ctx, collectionName)
		if err != nil {
			log.Fatalf("인덱스 \"%v\" 설명에 실패했습니다: %v", idx.Name, err)
		}

		idxConnection, err := r.Client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
		if err != nil {
			log.Fatalf("호스트에 대한 IndexConnection 생성에 실패했습니다: %v: %v", idx.Host, err)
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
		Dimension: vectorDim,
		Metric:    pinecone.Cosine,
		Cloud:     pinecone.Aws,
		Region:    "us-east-1",
	})

	if err != nil {
		return fmt.Errorf("Pinecone 인덱스 생성에 실패했습니다: %w", err)
	}
	idxConnection, err := r.Client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		log.Fatalf("호스트에 대한 IndexConnection 생성에 실패했습니다: %v: %v", idx.Host, err)
	}
	r.idxConnection = idxConnection
	log.Printf("Pinecone 인덱스 '%s'를 차원 %d로 생성했습니다", collectionName, vectorDim)
	return nil
}

func (r *PineconeRepo) InsertEmbedding(ctx context.Context, collectionName string, id string, embedding []float32) error {
	metadataMap := map[string]interface{}{
		"genre": "classical",
	}
	metadata, err := structpb.NewStruct(metadataMap)
	if err != nil {
		return fmt.Errorf("메타데이터 생성에 실패했습니다: %w", err)
	}

	vectors := []*pinecone.Vector{
		{
			Id:       id,
			Values:   embedding,
			Metadata: metadata,
		},
	}

	if _, err := r.idxConnection.UpsertVectors(ctx, vectors); err != nil {
		return fmt.Errorf("벡터 삽입에 실패했습니다: %w", err)
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
		log.Fatalf("메타데이터 필터 생성에 실패했습니다: %v", err)
	}

	res, err := r.idxConnection.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:         searchVector,
		TopK:           3,
		MetadataFilter: metadataFilter,
		IncludeValues:  true,
	})
	if err != nil {
		log.Fatalf("벡터로 쿼리하는 동안 오류가 발생했습니다: %v", err)
	}
	var ids []int64
	for _, match := range res.Matches {
		id, err := strconv.ParseInt(match.Vector.Id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ID 구문 분석에 실패했습니다: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Close는 Pinecone 클라이언트와 관련된 리소스를 정리합니다.
func (r *PineconeRepo) Close() error {
	log.Println("PineconeRepo.Close()가 호출되었습니다 - 닫을 리소스가 없습니다.")
	return nil
}
