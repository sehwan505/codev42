package repo

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"codev42/services/agent/storage"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type MilvusRepo struct {
	milvusConn *storage.MilvusConnection
}

func NewMilvusRepo(milvusConn *storage.MilvusConnection) *MilvusRepo {
	return &MilvusRepo{milvusConn: milvusConn}
}

func (r *MilvusRepo) InitCollection(ctx context.Context, collectionName string, vectorDim int) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	has, err := r.milvusConn.Client.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check milvus collection: %w", err)
	}
	if has {
		err = r.milvusConn.Client.LoadCollection(ctx, collectionName, false)
		if err != nil {
			return fmt.Errorf("failed to load milvus collection: %w", err)
		}
		return nil
	}

	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "Collection for storing code/file/function embeddings",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     false,
			},
			{
				Name:       "embeddings",
				DataType:   entity.FieldTypeFloatVector,
				TypeParams: map[string]string{"dim": fmt.Sprintf("%d", vectorDim)},
			},
		},
	}

	// Create the collection
	err = r.milvusConn.Client.CreateCollection(ctx, schema, 2)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Load the newly created collection
	err = r.milvusConn.Client.LoadCollection(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("failed to load milvus collection: %w", err)
	}

	return nil
}

// InsertEmbedding implements VectorDB.InsertEmbedding
func (r *MilvusRepo) InsertEmbedding(ctx context.Context, collectionName string, id string, embedding []float32) error {
	id_int64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse id: %w", err)
	}
	ids := entity.NewColumnInt64("id", []int64{id_int64})
	embeddings := entity.NewColumnFloatVector("embeddings", len(embedding), [][]float32{embedding})

	if _, err := r.milvusConn.Client.Insert(ctx, collectionName, "", ids, embeddings); err != nil {
		return fmt.Errorf("failed to insert vector into milvus: %w", err)
	}

	// Flush
	err = r.milvusConn.Client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("failed to flush milvus data: %w", err)
	}

	return nil
}

// SearchByVector implements VectorDB.SearchByVector
func (r *MilvusRepo) SearchByVector(ctx context.Context, collectionName string, searchVector []float32, topK int) ([]int64, error) {
	// Search with Flat index param (or IVF, HNSW param, etc. for advanced usage)
	sp, _ := entity.NewIndexFlatSearchParam()

	results, err := r.milvusConn.Client.Search(
		ctx,
		collectionName,
		[]string{},     // partition names, if any
		"",             // expression
		[]string{"id"}, // output fields
		[]entity.Vector{
			entity.FloatVector(searchVector),
		},
		"embeddings",
		entity.L2,
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search milvus: %w", err)
	}

	var ids []int64
	for _, result := range results {
		var idColumn *entity.ColumnInt64
		for _, field := range result.Fields {
			if field.Name() == "id" {
				c, ok := field.(*entity.ColumnInt64)
				if ok {
					idColumn = c
				}
			}
		}
		if idColumn == nil {
			log.Fatal("failed to find 'id' in result fields")
		}
		for i := 0; i < result.ResultCount; i++ {
			id, err := idColumn.ValueByIdx(i)
			if err != nil {
				log.Fatal(err.Error())
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// Close implements VectorDB.Close
func (r *MilvusRepo) Close() error {
	if r.milvusConn != nil {
		return r.milvusConn.Close()
	}
	return nil
}
