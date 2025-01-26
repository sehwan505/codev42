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

func NewFileRepo(dbConn *storage.RDBConnection) *FileRepo {
	return &FileRepo{dbConn: dbConn}
}

func (r *FileRepo) InsertFile(ctx context.Context, f *model.File) error {
	return r.dbConn.DB.WithContext(ctx).Create(f).Error
}
func (r *FileRepo) GetFileByPath(ctx context.Context, filePath string) (*model.File, error) {
	var file model.File
	err := r.dbConn.DB.WithContext(ctx).First(&file, filePath).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find file: %w", err)
	}
	return &file, nil
}

func (r *FileRepo) UpdateFile(ctx context.Context, f *model.File) error {
	return r.dbConn.DB.WithContext(ctx).Save(f).Error
}
func (r *FileRepo) DeleteFile(ctx context.Context, fileID int64) error {
	return r.dbConn.DB.WithContext(ctx).Delete(&model.File{}, fileID).Error
}
