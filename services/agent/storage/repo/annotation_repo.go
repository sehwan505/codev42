package repo

import (
	"context"

	"codev42-agent/model"
	"codev42-agent/storage"
)

// AnnotationRepository defines operations for Annotation entity
type AnnotationRepository interface {
	// CreateAnnotation creates a new Annotation
	CreateAnnotation(ctx context.Context, annotation *model.Annotation) error

	// UpdateAnnotation updates an existing Annotation
	UpdateAnnotation(ctx context.Context, annotation *model.Annotation) error

	// GetAnnotationByID retrieves an Annotation by its ID
	GetAnnotationByID(ctx context.Context, id int64) (*model.Annotation, error)

	// GetAnnotationsByPlanID retrieves all Annotations for a Plan
	GetAnnotationsByPlanID(ctx context.Context, planID int64) ([]model.Annotation, error)

	// DeleteAnnotation deletes an Annotation by its ID
	DeleteAnnotation(ctx context.Context, id int64) error

	// DeleteAnnotationsByPlanID deletes all Annotations for a Plan
	DeleteAnnotationsByPlanID(ctx context.Context, planID int64) error
}

// AnnotationRepo is the implementation of AnnotationRepository
type AnnotationRepo struct {
	dbConn *storage.RDBConnection
}

// NewAnnotationRepository creates a new AnnotationRepository
func NewAnnotationRepository(dbConn *storage.RDBConnection) AnnotationRepository {
	return &AnnotationRepo{dbConn: dbConn}
}

// CreateAnnotation creates a new Annotation
func (r *AnnotationRepo) CreateAnnotation(ctx context.Context, annotation *model.Annotation) error {
	return r.dbConn.DB.WithContext(ctx).Omit("id").Create(annotation).Error
}

// UpdateAnnotation updates an existing Annotation
func (r *AnnotationRepo) UpdateAnnotation(ctx context.Context, annotation *model.Annotation) error {
	return r.dbConn.DB.WithContext(ctx).Save(annotation).Error
}

// GetAnnotationByID retrieves an Annotation by its ID
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

// GetAnnotationsByPlanID retrieves all Annotations for a Plan
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

// DeleteAnnotation deletes an Annotation by its ID
func (r *AnnotationRepo) DeleteAnnotation(ctx context.Context, id int64) error {
	return r.dbConn.DB.WithContext(ctx).
		Delete(&model.Annotation{}, id).Error
}

// DeleteAnnotationsByPlanID deletes all Annotations for a Plan
func (r *AnnotationRepo) DeleteAnnotationsByPlanID(ctx context.Context, planID int64) error {
	return r.dbConn.DB.WithContext(ctx).
		Where("plan_id = ?", planID).
		Delete(&model.Annotation{}).Error
} 