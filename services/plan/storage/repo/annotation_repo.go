package repo

import (
	"context"

	"codev42-plan/model"
	"codev42-plan/storage"
)

// AnnotationRepository는 Annotation 엔티티에 대한 작업을 정의합니다.
type AnnotationRepository interface {
	// CreateAnnotation는 새로운 Annotation을 생성합니다.
	CreateAnnotation(ctx context.Context, annotation *model.Annotation) error

	// UpdateAnnotation는 기존 Annotation을 업데이트합니다.
	UpdateAnnotation(ctx context.Context, annotation *model.Annotation) error

	// GetAnnotationByID는 ID로 Annotation을 조회합니다.
	GetAnnotationByID(ctx context.Context, id int64) (*model.Annotation, error)

	// GetAnnotationsByPlanID는 Plan에 대한 모든 Annotation을 조회합니다.
	GetAnnotationsByPlanID(ctx context.Context, planID int64) ([]model.Annotation, error)

	// DeleteAnnotation는 ID로 Annotation을 삭제합니다.
	DeleteAnnotation(ctx context.Context, id int64) error

	// DeleteAnnotationsByPlanID는 Plan에 대한 모든 Annotation을 삭제합니다.
	DeleteAnnotationsByPlanID(ctx context.Context, planID int64) error
}

// AnnotationRepo는 AnnotationRepository의 구현체입니다.
type AnnotationRepo struct {
	dbConn *storage.RDBConnection
}

// NewAnnotationRepository는 새로운 AnnotationRepository를 생성합니다.
func NewAnnotationRepository(dbConn *storage.RDBConnection) AnnotationRepository {
	return &AnnotationRepo{dbConn: dbConn}
}

// CreateAnnotation는 새로운 Annotation을 생성합니다.
func (r *AnnotationRepo) CreateAnnotation(ctx context.Context, annotation *model.Annotation) error {
	return r.dbConn.DB.WithContext(ctx).Omit("id").Create(annotation).Error
}

// UpdateAnnotation는 기존 Annotation을 업데이트합니다.
func (r *AnnotationRepo) UpdateAnnotation(ctx context.Context, annotation *model.Annotation) error {
	return r.dbConn.DB.WithContext(ctx).Save(annotation).Error
}

// GetAnnotationByID는 ID로 Annotation을 조회합니다.
func (r *AnnotationRepo) GetAnnotationByID(ctx context.Context, id int64) (*model.Annotation, error) {
	var annotation model.Annotation
	err := r.dbConn.DB.WithContext(ctx).
		Where("id = ?", id).
		First(&annotation).Error

	if err != nil {
		return nil, err
	}

	return &annotation, nil
}

// GetAnnotationsByPlanID는 Plan에 대한 모든 Annotation을 조회합니다.
func (r *AnnotationRepo) GetAnnotationsByPlanID(ctx context.Context, planID int64) ([]model.Annotation, error) {
	var annotations []model.Annotation
	err := r.dbConn.DB.WithContext(ctx).
		Where("plan_id = ?", planID).
		Find(&annotations).Error

	if err != nil {
		return nil, err
	}

	return annotations, nil
}

// DeleteAnnotation는 ID로 Annotation을 삭제합니다.
func (r *AnnotationRepo) DeleteAnnotation(ctx context.Context, id int64) error {
	return r.dbConn.DB.WithContext(ctx).
		Delete(&model.Annotation{}, id).Error
}

// DeleteAnnotationsByPlanID는 Plan에 대한 모든 Annotation을 삭제합니다.
func (r *AnnotationRepo) DeleteAnnotationsByPlanID(ctx context.Context, planID int64) error {
	return r.dbConn.DB.WithContext(ctx).
		Where("plan_id = ?", planID).
		Delete(&model.Annotation{}).Error
}
