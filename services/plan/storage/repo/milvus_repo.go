package repo

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"codev42-plan/storage"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type MilvusRepo struct {
	milvusConn *storage.MilvusConnection
}

func NewMilvusRepo(milvusConn *storage.MilvusConnection) *MilvusRepo {
	return &MilvusRepo{milvusConn: milvusConn}
}

func (r *MilvusRepo) InitCollection(ctx context.Context, collectionName string, vectorDim int32) error {
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

	// 컬렉션 생성
	err = r.milvusConn.Client.CreateCollection(ctx, schema, 2)
	if err != nil {
		return fmt.Errorf("컬렉션 생성 실패: %w", err)
	}

	// 새로 생성된 컬렉션 로드
	err = r.milvusConn.Client.LoadCollection(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("milvus 컬렉션 로드 실패: %w", err)
	}

	return nil
}

// InsertEmbedding은 VectorDB.InsertEmbedding을 구현합니다.
func (r *MilvusRepo) InsertEmbedding(ctx context.Context, collectionName string, id string, embedding []float32) error {
	id_int64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("ID 파싱 실패: %w", err)
	}
	ids := entity.NewColumnInt64("id", []int64{id_int64})
	embeddings := entity.NewColumnFloatVector("embeddings", len(embedding), [][]float32{embedding})

	if _, err := r.milvusConn.Client.Insert(ctx, collectionName, "", ids, embeddings); err != nil {
		return fmt.Errorf("milvus에 벡터 삽입 실패: %w", err)
	}

	// 플러시
	err = r.milvusConn.Client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("milvus 데이터 플러시 실패: %w", err)
	}

	return nil
}

// SearchByVector는 VectorDB.SearchByVector를 구현합니다.
func (r *MilvusRepo) SearchByVector(ctx context.Context, collectionName string, searchVector []float32, topK int) ([]int64, error) {
	// Flat 인덱스 매개변수로 검색 (또는 고급 사용을 위해 IVF, HNSW 매개변수 등)
	sp, _ := entity.NewIndexFlatSearchParam()

	results, err := r.milvusConn.Client.Search(
		ctx,
		collectionName,
		[]string{},     // 파티션 이름 (있는 경우)
		"",             // 표현식
		[]string{"id"}, // 출력 필드
		[]entity.Vector{
			entity.FloatVector(searchVector),
		},
		"embeddings",
		entity.L2,
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("milvus 검색 실패: %w", err)
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
			log.Fatal("결과 필드에서 'id'를 찾지 못했습니다")
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

func (r *MilvusRepo) DeleteByID(ctx context.Context, collectionName string, id string) error {
	// if err := r.milvusConn.Client.Delete(ctx, collectionName, ids); err != nil {
	// 	return fmt.Errorf("벡터 삭제 실패: %w", err)
	// }
	// TODO: 구현 필요
	return nil
}

// Close는 VectorDB.Close를 구현합니다.
func (r *MilvusRepo) Close() error {
	if r.milvusConn != nil {
		return r.milvusConn.Close()
	}
	return nil
}
