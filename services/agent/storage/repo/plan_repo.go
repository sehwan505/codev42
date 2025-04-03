package repo

import (
	"context"

	"codev42-agent/model"
	"codev42-agent/storage"

	"gorm.io/gorm"
)

type PlanRepository interface {
	// CreateDevPlanWithDetails는 DevPlan, Plan, Annotation을 한 번에 생성합니다.
	CreateDevPlanWithDetails(ctx context.Context, devPlan *model.DevPlan) error

	// UpdateDevPlanWithDetails는 DevPlan과 관련 Plan, Annotation을 업데이트합니다.
	UpdateDevPlanWithDetails(ctx context.Context, devPlan *model.DevPlan) error

	// GetDevPlanByID는 ID로 DevPlan을 조회합니다.
	GetDevPlanByID(ctx context.Context, id int64) (*model.DevPlan, error)

	// GetDevPlansByProjectID는 프로젝트 ID로 DevPlan 목록을 조회합니다.
	GetDevPlansByProjectID(ctx context.Context, projectID string) ([]model.DevPlan, error)

	// DeleteDevPlan은 DevPlan과 관련 Plan, Annotation을 삭제합니다.
	DeleteDevPlan(ctx context.Context, id int64) error
}

type PlanRepo struct {
	dbConn *storage.RDBConnection
}

// NewPlanRepository는 새로운 PlanRepository 인스턴스를 생성합니다.
func NewPlanRepository(dbConn *storage.RDBConnection) PlanRepository {
	return &PlanRepo{dbConn: dbConn}
}

// CreateDevPlanWithDetails는 DevPlan, Plan, Annotation을 한 번에 생성합니다.
func (r *PlanRepo) CreateDevPlanWithDetails(ctx context.Context, devPlan *model.DevPlan) error {
	return r.dbConn.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(devPlan).Error; err != nil {
			return err
		}

		// Plans 생성 시 ID 필드를 제외
		for i := range devPlan.Plans {
			devPlan.Plans[i].DevPlanID = devPlan.ID
			if err := tx.Omit("id").Create(&devPlan.Plans[i]).Error; err != nil {
				return err
			}

			// Annotations 생성
			for j := range devPlan.Plans[i].Annotations {
				devPlan.Plans[i].Annotations[j].PlanID = devPlan.Plans[i].ID
				if err := tx.Omit("id").Create(&devPlan.Plans[i].Annotations[j]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// UpdateDevPlanWithDetails는 DevPlan과 관련 Plan, Annotation을 업데이트합니다.
func (r *PlanRepo) UpdateDevPlanWithDetails(ctx context.Context, devPlan *model.DevPlan) error {
	return r.dbConn.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(devPlan).Error; err != nil {
			return err
		}

		var existingPlans []model.Plan
		if err := tx.Where("dev_plan_id = ?", devPlan.ID).Find(&existingPlans).Error; err != nil {
			return err
		}

		existingPlanIDs := make(map[int64]bool)
		for _, plan := range existingPlans {
			existingPlanIDs[plan.ID] = true
		}

		for i := range devPlan.Plans {
			devPlan.Plans[i].DevPlanID = devPlan.ID

			if devPlan.Plans[i].ID > 0 && existingPlanIDs[devPlan.Plans[i].ID] {
				if err := tx.Save(&devPlan.Plans[i]).Error; err != nil {
					return err
				}

				var existingAnnotations []model.Annotation
				if err := tx.Where("plan_id = ?", devPlan.Plans[i].ID).Find(&existingAnnotations).Error; err != nil {
					return err
				}

				existingAnnotationIDs := make(map[int64]bool)
				for _, annotation := range existingAnnotations {
					existingAnnotationIDs[annotation.ID] = true
				}

				for j := range devPlan.Plans[i].Annotations {
					devPlan.Plans[i].Annotations[j].PlanID = devPlan.Plans[i].ID

					if devPlan.Plans[i].Annotations[j].ID > 0 && existingAnnotationIDs[devPlan.Plans[i].Annotations[j].ID] {
						if err := tx.Save(&devPlan.Plans[i].Annotations[j]).Error; err != nil {
							return err
						}

						delete(existingAnnotationIDs, devPlan.Plans[i].Annotations[j].ID)
					} else {
						if err := tx.Create(&devPlan.Plans[i].Annotations[j]).Error; err != nil {
							return err
						}
					}
				}

				for annotationID := range existingAnnotationIDs {
					if err := tx.Delete(&model.Annotation{}, annotationID).Error; err != nil {
						return err
					}
				}

				delete(existingPlanIDs, devPlan.Plans[i].ID)
			} else {
				if err := tx.Create(&devPlan.Plans[i]).Error; err != nil {
					return err
				}

				for j := range devPlan.Plans[i].Annotations {
					devPlan.Plans[i].Annotations[j].PlanID = devPlan.Plans[i].ID
					if err := tx.Create(&devPlan.Plans[i].Annotations[j]).Error; err != nil {
						return err
					}
				}
			}
		}

		for planID := range existingPlanIDs {
			if err := tx.Delete(&model.Plan{}, planID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetDevPlanByID는 ID로 DevPlan을 조회합니다.
func (r *PlanRepo) GetDevPlanByID(ctx context.Context, id int64) (*model.DevPlan, error) {
	var devPlan model.DevPlan
	err := r.dbConn.DB.WithContext(ctx).
		Preload("Plans").
		Preload("Plans.Annotations").
		Where("id = ?", id).
		First(&devPlan).Error

	if err != nil {
		return nil, err
	}

	return &devPlan, nil
}

// GetDevPlansByProjectID는 프로젝트 ID로 DevPlan 목록을 조회합니다.
func (r *PlanRepo) GetDevPlansByProjectID(ctx context.Context, projectID string) ([]model.DevPlan, error) {
	var devPlans []model.DevPlan
	err := r.dbConn.DB.WithContext(ctx).
		Preload("Plans").
		Preload("Plans.Annotations").
		Where("project_id = ?", projectID).
		Find(&devPlans).Error

	if err != nil {
		return nil, err
	}

	return devPlans, nil
}

// DeleteDevPlan은 DevPlan과 관련 Plan, Annotation을 삭제합니다.
func (r *PlanRepo) DeleteDevPlan(ctx context.Context, id int64) error {
	return r.dbConn.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var plans []model.Plan
		if err := tx.Where("dev_plan_id = ?", id).Find(&plans).Error; err != nil {
			return err
		}

		for _, plan := range plans {
			if err := tx.Where("plan_id = ?", plan.ID).Delete(&model.Annotation{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("dev_plan_id = ?", id).Delete(&model.Plan{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.DevPlan{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}
