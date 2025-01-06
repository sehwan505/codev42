package repo

import (
	"context"
	"fmt"

	"codev42/services/agent/model"
	"codev42/services/agent/storage"
)

// CodeRepo : Code 엔티티에 대한 MySQL Repo
type CodeRepo struct {
	dbConn *storage.DBConnection
}

// NewCodeRepo : CodeRepo 생성
func NewCodeRepo(dbConn *storage.DBConnection) *CodeRepo {
	return &CodeRepo{dbConn: dbConn}
}

// InsertCode : 함수 정보 삽입
func (r *CodeRepo) InsertCode(ctx context.Context, fn *model.Code) error {
	return r.dbConn.DB.WithContext(ctx).Create(fn).Error
}

// GetCodeByID : 함수 단건 조회
func (r *CodeRepo) GetCodeByID(ctx context.Context, fnID int64) (*model.Code, error) {
	var fn model.Code
	err := r.dbConn.DB.WithContext(ctx).First(&fn, fnID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find Code: %w", err)
	}
	return &fn, nil
}

// UpdateCode : 함수 수정
func (r *CodeRepo) UpdateCode(ctx context.Context, fn *model.Code) error {
	return r.dbConn.DB.WithContext(ctx).Save(fn).Error
}

// DeleteCode : 함수 삭제
func (r *CodeRepo) DeleteCode(ctx context.Context, fnID int64) error {
	return r.dbConn.DB.WithContext(ctx).Delete(&model.Code{}, fnID).Error
}
