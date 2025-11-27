package repo

import (
	"context"
	"fmt"

	"codev42-plan/model"
	"codev42-plan/storage"
)

// FileRepo : File 엔티티에 대한 MySQL Repo
type FileRepo struct {
	dbConn *storage.RDBConnection
}

func NewFileRepo(dbConn *storage.RDBConnection) *FileRepo {
	return &FileRepo{dbConn: dbConn}
}

func (r *FileRepo) InsertFile(ctx context.Context, f *model.File) (int64, error) {
	if f.FilePath == "" {
		return 0, fmt.Errorf("file path is required")
	}
	err := r.dbConn.DB.WithContext(ctx).Create(f).Error
	if err != nil {
		return 0, err
	}
	return f.ID, nil
}

func (r *FileRepo) GetFileByPath(ctx context.Context, filePath string) (*model.File, error) {
	var file model.File
	err := r.dbConn.DB.WithContext(ctx).Where("file_path = ?", filePath).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepo) UpdateFile(ctx context.Context, f *model.File) error {
	return r.dbConn.DB.WithContext(ctx).Save(f).Error
}
func (r *FileRepo) DeleteFile(ctx context.Context, fileID int64) error {
	return r.dbConn.DB.WithContext(ctx).Delete(&model.File{}, fileID).Error
}
