package repo

import (
	"context"

	"codev42-plan/model"
	"codev42-plan/storage"
)

// DevPlanRepository는 DevPlan 엔티티에 대한 작업을 정의합니다.
type DevPlanRepository interface {
	// CreateDevPlan는 새로운 DevPlan을 생성합니다.
	CreateDevPlan(ctx context.Context, devPlan *model.DevPlan) error

	// UpdateDevPlan는 기존 DevPlan을 업데이트합니다.
	UpdateDevPlan(ctx context.Context, devPlan *model.DevPlan) error

	// GetDevPlanByID는 ID로 DevPlan을 조회합니다.
	GetDevPlanByID(ctx context.Context, id int64) (*model.DevPlan, error)

	// GetDevPlansByProjectID는 프로젝트의 모든 DevPlan을 조회합니다.
	GetDevPlansByProjectID(ctx context.Context, projectID string, branch string) ([]DevPlanListElement, error)

	// DeleteDevPlan는 ID로 DevPlan을 삭제합니다.
	DeleteDevPlan(ctx context.Context, id int64) error
}

type DevPlanListElement struct {
	ID     int64  `json:"id"`
	Prompt string `json:"prompt"`
}

// DevPlanRepo는 DevPlanRepository의 구현체입니다.
type DevPlanRepo struct {
	dbConn *storage.RDBConnection
}

// NewDevPlanRepository는 새로운 DevPlanRepository를 생성합니다.
func NewDevPlanRepository(dbConn *storage.RDBConnection) DevPlanRepository {
	return &DevPlanRepo{dbConn: dbConn}
}

// CreateDevPlan는 새로운 DevPlan을 생성합니다.
func (r *DevPlanRepo) CreateDevPlan(ctx context.Context, devPlan *model.DevPlan) error {
	return r.dbConn.DB.WithContext(ctx).Create(devPlan).Error
}

// UpdateDevPlan는 기존 DevPlan을 업데이트합니다.
func (r *DevPlanRepo) UpdateDevPlan(ctx context.Context, devPlan *model.DevPlan) error {
	return r.dbConn.DB.WithContext(ctx).Save(devPlan).Error
}

// GetDevPlanByID는 ID로 DevPlan을 조회합니다.
func (r *DevPlanRepo) GetDevPlanByID(ctx context.Context, id int64) (*model.DevPlan, error) {
	var devPlan model.DevPlan
	err := r.dbConn.DB.WithContext(ctx).
		Where("id = ?", id).
		First(&devPlan).Error

	if err != nil {
		return nil, err
	}

	return &devPlan, nil
}

// GetDevPlansByProjectID는 프로젝트의 모든 DevPlan을 조회합니다.
func (r *DevPlanRepo) GetDevPlansByProjectID(ctx context.Context, projectID string, branch string) ([]DevPlanListElement, error) {
	var devPlanList []DevPlanListElement
	err := r.dbConn.DB.WithContext(ctx).
		Model(&model.DevPlan{}).
		Select("id", "prompt").
		Where("project_id = ? AND branch = ?", projectID, branch).
		Find(&devPlanList).Error

	if err != nil {
		return nil, err
	}
	return devPlanList, nil
}

// DeleteDevPlan는 ID로 DevPlan을 삭제합니다.
func (r *DevPlanRepo) DeleteDevPlan(ctx context.Context, id int64) error {
	return r.dbConn.DB.WithContext(ctx).
		Delete(&model.DevPlan{}, id).Error
}
