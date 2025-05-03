package service

import (
	"context"

	"codev42-agent/model"
	"codev42-agent/storage/repo"
)

// PlanService coordinates operations between DevPlanRepo, PlanRepo, and AnnotationRepo
type PlanService struct {
	devPlanRepo    repo.DevPlanRepository
	planRepo       repo.PlanRepository
	annotationRepo repo.AnnotationRepository
}

// NewPlanService creates a new PlanService instance
func NewPlanService(
	devPlanRepo repo.DevPlanRepository,
	planRepo repo.PlanRepository,
	annotationRepo repo.AnnotationRepository,
) *PlanService {
	return &PlanService{
		devPlanRepo:    devPlanRepo,
		planRepo:       planRepo,
		annotationRepo: annotationRepo,
	}
}

func (s *PlanService) CreateDevPlanWithDetails(ctx context.Context, devPlan *model.DevPlan) error {
	if err := s.devPlanRepo.CreateDevPlan(ctx, devPlan); err != nil {
		return err
	}

	for i := range devPlan.Plans {
		devPlan.Plans[i].DevPlanID = devPlan.ID
		if err := s.planRepo.CreatePlan(ctx, &devPlan.Plans[i]); err != nil {
			return err
		}

		for j := range devPlan.Plans[i].Annotations {
			devPlan.Plans[i].Annotations[j].PlanID = devPlan.Plans[i].ID
			if err := s.annotationRepo.CreateAnnotation(ctx, &devPlan.Plans[i].Annotations[j]); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *PlanService) UpdateDevPlanWithDetails(ctx context.Context, devPlan *model.DevPlan) error {
	if err := s.devPlanRepo.UpdateDevPlan(ctx, devPlan); err != nil {
		return err
	}

	existingPlans, err := s.planRepo.GetPlansByDevPlanID(ctx, devPlan.ID)
	if err != nil {
		return err
	}

	existingPlanIDs := make(map[int64]*model.Plan)
	for i, plan := range existingPlans {
		existingPlanIDs[plan.ID] = &existingPlans[i]
	}

	// 3. Process each Plan in the updated DevPlan
	for i := range devPlan.Plans {
		devPlan.Plans[i].DevPlanID = devPlan.ID

		if devPlan.Plans[i].ID > 0 && existingPlanIDs[devPlan.Plans[i].ID] != nil {
			// Update existing Plan
			if err := s.planRepo.UpdatePlan(ctx, &devPlan.Plans[i]); err != nil {
				return err
			}

			// Get existing Annotations for this Plan
			existingAnnotations, err := s.annotationRepo.GetAnnotationsByPlanID(ctx, devPlan.Plans[i].ID)
			if err != nil {
				return err
			}

			// Make a map of existing Annotation IDs
			existingAnnotationIDs := make(map[int64]*model.Annotation)
			for j, annotation := range existingAnnotations {
				existingAnnotationIDs[annotation.ID] = &existingAnnotations[j]
			}

			// Process each Annotation
			for j := range devPlan.Plans[i].Annotations {
				devPlan.Plans[i].Annotations[j].PlanID = devPlan.Plans[i].ID

				if devPlan.Plans[i].Annotations[j].ID > 0 && existingAnnotationIDs[devPlan.Plans[i].Annotations[j].ID] != nil {
					// Update existing Annotation
					if err := s.annotationRepo.UpdateAnnotation(ctx, &devPlan.Plans[i].Annotations[j]); err != nil {
						return err
					}
					// Remove from map to track which ones to delete later
					delete(existingAnnotationIDs, devPlan.Plans[i].Annotations[j].ID)
				} else {
					// Create new Annotation
					if err := s.annotationRepo.CreateAnnotation(ctx, &devPlan.Plans[i].Annotations[j]); err != nil {
						return err
					}
				}
			}

			// Delete Annotations that were not in the update
			for annotationID := range existingAnnotationIDs {
				if err := s.annotationRepo.DeleteAnnotation(ctx, annotationID); err != nil {
					return err
				}
			}

			// Remove from map to track which Plans to delete later
			delete(existingPlanIDs, devPlan.Plans[i].ID)
		} else {
			// Create new Plan
			if err := s.planRepo.CreatePlan(ctx, &devPlan.Plans[i]); err != nil {
				return err
			}

			// Create Annotations for the new Plan
			for j := range devPlan.Plans[i].Annotations {
				devPlan.Plans[i].Annotations[j].PlanID = devPlan.Plans[i].ID
				if err := s.annotationRepo.CreateAnnotation(ctx, &devPlan.Plans[i].Annotations[j]); err != nil {
					return err
				}
			}
		}
	}

	// Delete Plans that were not in the update
	for planID := range existingPlanIDs {
		if err := s.annotationRepo.DeleteAnnotationsByPlanID(ctx, planID); err != nil {
			return err
		}
		if err := s.planRepo.DeletePlan(ctx, planID); err != nil {
			return err
		}
	}

	return nil
}

// GetDevPlanByID retrieves a DevPlan with its Plans and Annotations
func (s *PlanService) GetDevPlanByID(ctx context.Context, id int64) (*model.DevPlan, error) {
	// 1. Get DevPlan
	devPlan, err := s.devPlanRepo.GetDevPlanByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Get Plans for this DevPlan
	plans, err := s.planRepo.GetPlansByDevPlanID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Get Annotations for each Plan and build the complete structure
	for i := range plans {
		annotations, err := s.annotationRepo.GetAnnotationsByPlanID(ctx, plans[i].ID)
		if err != nil {
			return nil, err
		}
		plans[i].Annotations = annotations
	}
	devPlan.Plans = plans

	return devPlan, nil
}

func (s *PlanService) GetDevPlansByProjectID(ctx context.Context, projectID string, branch string) ([]repo.DevPlanListElement, error) {
	devPlans, err := s.devPlanRepo.GetDevPlansByProjectID(ctx, projectID, branch)
	if err != nil {
		return nil, err
	}

	return devPlans, nil
}

// DeleteDevPlan deletes a DevPlan and all associated Plans and Annotations
func (s *PlanService) DeleteDevPlan(ctx context.Context, id int64) error {
	// 1. Get Plans for this DevPlan
	plans, err := s.planRepo.GetPlansByDevPlanID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Delete all Annotations for each Plan
	for _, plan := range plans {
		if err := s.annotationRepo.DeleteAnnotationsByPlanID(ctx, plan.ID); err != nil {
			return err
		}
	}

	// 3. Delete all Plans for this DevPlan
	if err := s.planRepo.DeletePlansByDevPlanID(ctx, id); err != nil {
		return err
	}

	// 4. Delete the DevPlan
	return s.devPlanRepo.DeleteDevPlan(ctx, id)
}
