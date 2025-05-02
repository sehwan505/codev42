package repo

import (
	"context"

	"codev42-agent/model"
	"codev42-agent/storage"
)

// PlanRepository defines operations for Plan entity
type PlanRepository interface {
	// CreatePlan creates a new Plan
	CreatePlan(ctx context.Context, plan *model.Plan) error

	// UpdatePlan updates an existing Plan
	UpdatePlan(ctx context.Context, plan *model.Plan) error

	// GetPlanByID retrieves a Plan by its ID
	GetPlanByID(ctx context.Context, id int64) (*model.Plan, error)

	// GetPlansByDevPlanID retrieves all Plans for a DevPlan
	GetPlansByDevPlanID(ctx context.Context, devPlanID int64) ([]model.Plan, error)

	// DeletePlan deletes a Plan by its ID
	DeletePlan(ctx context.Context, id int64) error

	// DeletePlansByDevPlanID deletes all Plans for a DevPlan
	DeletePlansByDevPlanID(ctx context.Context, devPlanID int64) error
}

// PlanEntityRepo is the implementation of PlanRepository
type PlanEntityRepo struct {
	dbConn *storage.RDBConnection
}

// NewPlanRepository creates a new PlanRepository
func NewPlanRepository(dbConn *storage.RDBConnection) PlanRepository {
	return &PlanEntityRepo{dbConn: dbConn}
}

// CreatePlan creates a new Plan
func (r *PlanEntityRepo) CreatePlan(ctx context.Context, plan *model.Plan) error {
	return r.dbConn.DB.WithContext(ctx).Omit("id").Create(plan).Error
}

// UpdatePlan updates an existing Plan
func (r *PlanEntityRepo) UpdatePlan(ctx context.Context, plan *model.Plan) error {
	return r.dbConn.DB.WithContext(ctx).Save(plan).Error
}

// GetPlanByID retrieves a Plan by its ID
func (r *PlanEntityRepo) GetPlanByID(ctx context.Context, id int64) (*model.Plan, error) {
	var plan model.Plan
	err := r.dbConn.DB.WithContext(ctx).
		Where("id = ?", id).
		First(&plan).Error

	if err != nil {
		return nil, err
	}

	return &plan, nil
}

// GetPlansByDevPlanID retrieves all Plans for a DevPlan
func (r *PlanEntityRepo) GetPlansByDevPlanID(ctx context.Context, devPlanID int64) ([]model.Plan, error) {
	var plans []model.Plan
	err := r.dbConn.DB.WithContext(ctx).
		Where("dev_plan_id = ?", devPlanID).
		Find(&plans).Error

	if err != nil {
		return nil, err
	}

	return plans, nil
}

// DeletePlan deletes a Plan by its ID
func (r *PlanEntityRepo) DeletePlan(ctx context.Context, id int64) error {
	return r.dbConn.DB.WithContext(ctx).
		Delete(&model.Plan{}, id).Error
}

// DeletePlansByDevPlanID deletes all Plans for a DevPlan
func (r *PlanEntityRepo) DeletePlansByDevPlanID(ctx context.Context, devPlanID int64) error {
	return r.dbConn.DB.WithContext(ctx).
		Where("dev_plan_id = ?", devPlanID).
		Delete(&model.Plan{}).Error
} 