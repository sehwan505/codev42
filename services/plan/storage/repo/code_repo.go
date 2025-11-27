package repo

import (
	"context"
	"fmt"

	"codev42-plan/model"
	"codev42-plan/storage"
)

// CodeRepo : Code 엔티티에 대한 MySQL Repo
type CodeRepo struct {
	dbConn *storage.RDBConnection
}

// NewCodeRepo : CodeRepo 생성
func NewCodeRepo(dbConn *storage.RDBConnection) *CodeRepo {
	return &CodeRepo{dbConn: dbConn}
}

// InsertCode : 함수 정보 삽입
func (r *CodeRepo) InsertCode(ctx context.Context, fn *model.Code) (int64, error) {
	err := r.dbConn.DB.WithContext(ctx).Create(fn).Error
	if err != nil {
		return 0, err
	}
	return fn.ID, nil
}

// GetCodeByID : 함수 단건 조회
func (r *CodeRepo) GetCodeByFileIdAndName(ctx context.Context, fileID int64, funcDeclaration string) (*model.Code, error) {
	var fn model.Code
	err := r.dbConn.DB.WithContext(ctx).Where("file_id = ? AND func_declaration = ?", fileID, funcDeclaration).First(&fn).Error
	if err != nil {
		return nil, fmt.Errorf("record not found")
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
