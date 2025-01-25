package repo

import (
	"context"
	"fmt"

	"codev42/services/agent/model"
	"codev42/services/agent/storage"
)

// FileRepo : File 엔티티에 대한 MySQL Repo
type FileRepo struct {
	dbConn *storage.RDBConnection
}

// NewFileRepo : FileRepo 생성
func NewFileRepo(dbConn *storage.RDBConnection) *FileRepo {
	return &FileRepo{dbConn: dbConn}
}

// InsertFile : 파일 정보 삽입
func (r *FileRepo) InsertFile(ctx context.Context, f *model.FileStruct) error {
	return r.dbConn.DB.WithContext(ctx).Create(f).Error
}

// GetFileByID : 파일 단건 조회
func (r *FileRepo) GetFileByID(ctx context.Context, fileID int64) (*model.FileStruct, error) {
	var file model.FileStruct
	err := r.dbConn.DB.WithContext(ctx).Preload("Functions").First(&file, fileID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find file: %w", err)
	}
	return &file, nil
}

// UpdateFile : 파일 수정
func (r *FileRepo) UpdateFile(ctx context.Context, f *model.FileStruct) error {
	return r.dbConn.DB.WithContext(ctx).Save(f).Error
}

// DeleteFile : 파일 삭제
func (r *FileRepo) DeleteFile(ctx context.Context, fileID int64) error {
	return r.dbConn.DB.WithContext(ctx).Delete(&model.FileStruct{}, fileID).Error
}
