package repo

import (
	"context"
	"fmt"

	"codev42/services/agent/model"
	"codev42/services/agent/storage"
)

// ProjectRepo : Project 엔티티에 대한 MySQL Repo
type ProjectRepo struct {
	dbConn *storage.RDBConnection
}

// NewProjectRepo : ProjectRepo 생성
func NewProjectRepo(dbConn *storage.RDBConnection) *ProjectRepo {
	return &ProjectRepo{dbConn: dbConn}
}

// InsertProject : 프로젝트 저장
func (r *ProjectRepo) InsertProject(ctx context.Context, p *model.Project) error {
	return r.dbConn.DB.WithContext(ctx).Create(p).Error
}

// GetProjectByID : 프로젝트 단건 조회
func (r *ProjectRepo) GetProjectByID(ctx context.Context, projectID int64) (*model.Project, error) {
	var project model.Project
	err := r.dbConn.DB.WithContext(ctx).
		Preload("Files.Functions"). // 연관된 Files 및 Functions도 함께 로드
		First(&project, projectID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	return &project, nil
}

// ListProjects : 프로젝트 리스트 조회(간단 예시)
func (r *ProjectRepo) ListProjects(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	err := r.dbConn.DB.WithContext(ctx).Find(&projects).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}

// UpdateProject : 프로젝트 업데이트
func (r *ProjectRepo) UpdateProject(ctx context.Context, p *model.Project) error {
	return r.dbConn.DB.WithContext(ctx).Save(p).Error
}

// DeleteProject : 프로젝트 삭제
func (r *ProjectRepo) DeleteProject(ctx context.Context, projectID int64) error {
	return r.dbConn.DB.WithContext(ctx).Delete(&model.Project{}, projectID).Error
}
