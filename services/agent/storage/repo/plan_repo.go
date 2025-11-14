package repo

import (
	"context"

	"codev42-agent/model"
	"codev42-agent/storage"
)

// PlanRepository는 Plan 엔티티에 대한 작업을 정의합니다.
type PlanRepository interface {
	// CreatePlan는 새로운 Plan을 생성합니다.
	CreatePlan(ctx context.Context, plan *model.Plan) error

	// UpdatePlan는 기존 Plan을 업데이트합니다.
	UpdatePlan(ctx context.Context, plan *model.Plan) error

	// GetPlanByID는 ID로 Plan을 조회합니다.
	GetPlanByID(ctx context.Context, id int64) (*model.Plan, error)

	// GetPlansByDevPlanID는 DevPlan에 대한 모든 Plan을 조회합니다.
	GetPlansByDevPlanID(ctx context.Context, devPlanID int64) ([]model.Plan, error)

	// DeletePlan는 ID로 Plan을 삭제합니다.
	DeletePlan(ctx context.Context, id int64) error

	// DeletePlansByDevPlanID는 DevPlan에 대한 모든 Plan을 삭제합니다.
	DeletePlansByDevPlanID(ctx context.Context, devPlanID int64) error
}

// PlanEntityRepo는 PlanRepository의 구현체입니다.
type PlanEntityRepo struct {
	dbConn *storage.RDBConnection
}

// NewPlanRepository는 새로운 PlanRepository를 생성합니다.
func NewPlanRepository(dbConn *storage.RDBConnection) PlanRepository {
	return &PlanEntityRepo{dbConn: dbConn}
}

// CreatePlan는 새로운 Plan을 생성합니다.
func (r *PlanEntityRepo) CreatePlan(ctx context.Context, plan *model.Plan) error {
	return r.dbConn.DB.WithContext(ctx).Omit("id").Create(plan).Error
}

// UpdatePlan는 기존 Plan을 업데이트합니다.
func (r *PlanEntityRepo) UpdatePlan(ctx context.Context, plan *model.Plan) error {
	return r.dbConn.DB.WithContext(ctx).Save(plan).Error
}

// GetPlanByID는 ID로 Plan을 조회합니다.
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

// GetPlansByDevPlanID는 DevPlan에 대한 모든 Plan을 조회합니다.
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

// DeletePlan는 ID로 Plan을 삭제합니다.
func (r *PlanEntityRepo) DeletePlan(ctx context.Context, id int64) error {
	return r.dbConn.DB.WithContext(ctx).
		Delete(&model.Plan{}, id).Error
}

// DeletePlansByDevPlanID는 DevPlan에 대한 모든 Plan을 삭제합니다.
func (r *PlanEntityRepo) DeletePlansByDevPlanID(ctx context.Context, devPlanID int64) error {
	return r.dbConn.DB.WithContext(ctx).
		Where("dev_plan_id = ?", devPlanID).
		Delete(&model.Plan{}).Error
} 