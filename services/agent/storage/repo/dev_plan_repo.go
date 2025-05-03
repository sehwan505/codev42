package repo

import (
	"context"

	"codev42-agent/model"
	"codev42-agent/storage"
)

// DevPlanRepository defines operations for DevPlan entity
type DevPlanRepository interface {
	// CreateDevPlan creates a new DevPlan
	CreateDevPlan(ctx context.Context, devPlan *model.DevPlan) error

	// UpdateDevPlan updates an existing DevPlan
	UpdateDevPlan(ctx context.Context, devPlan *model.DevPlan) error

	// GetDevPlanByID retrieves a DevPlan by its ID
	GetDevPlanByID(ctx context.Context, id int64) (*model.DevPlan, error)

	// GetDevPlansByProjectID retrieves all DevPlans for a project
	GetDevPlansByProjectID(ctx context.Context, projectID string, branch string) ([]DevPlanListElement, error)

	// DeleteDevPlan deletes a DevPlan by its ID
	DeleteDevPlan(ctx context.Context, id int64) error
}

type DevPlanListElement struct {
	ID     int64  `json:"id"`
	Prompt string `json:"prompt"`
}

// DevPlanRepo is the implementation of DevPlanRepository
type DevPlanRepo struct {
	dbConn *storage.RDBConnection
}

// NewDevPlanRepository creates a new DevPlanRepository
func NewDevPlanRepository(dbConn *storage.RDBConnection) DevPlanRepository {
	return &DevPlanRepo{dbConn: dbConn}
}

// CreateDevPlan creates a new DevPlan
func (r *DevPlanRepo) CreateDevPlan(ctx context.Context, devPlan *model.DevPlan) error {
	return r.dbConn.DB.WithContext(ctx).Create(devPlan).Error
}

// UpdateDevPlan updates an existing DevPlan
func (r *DevPlanRepo) UpdateDevPlan(ctx context.Context, devPlan *model.DevPlan) error {
	return r.dbConn.DB.WithContext(ctx).Save(devPlan).Error
}

// GetDevPlanByID retrieves a DevPlan by its ID
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

// GetDevPlansByProjectID retrieves all DevPlans for a project
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

// DeleteDevPlan deletes a DevPlan by its ID
func (r *DevPlanRepo) DeleteDevPlan(ctx context.Context, id int64) error {
	return r.dbConn.DB.WithContext(ctx).
		Delete(&model.DevPlan{}, id).Error
}
