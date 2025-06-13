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

	// 1. 기존 Plan들을 가져옵니다
	existingPlans, err := s.planRepo.GetPlansByDevPlanID(ctx, devPlan.ID)
	if err != nil {
		return err
	}

	// 2. 기존 Plan들의 map을 만듭니다
	existingPlanIDs := make(map[int64]*model.Plan)
	for i := range existingPlans {
		existingPlanIDs[existingPlans[i].ID] = &existingPlans[i]
	}

	// 3. 새로운 Plan들을 처리합니다
	for i := range devPlan.Plans {
		devPlan.Plans[i].DevPlanID = devPlan.ID

		if devPlan.Plans[i].ID > 0 && existingPlanIDs[devPlan.Plans[i].ID] != nil {
			// 3.1 기존 Plan 업데이트
			if err := s.planRepo.UpdatePlan(ctx, &devPlan.Plans[i]); err != nil {
				return err
			}

			// 3.2 기존 Annotation들을 가져옵니다
			existingAnnotations, err := s.annotationRepo.GetAnnotationsByPlanID(ctx, devPlan.Plans[i].ID)
			if err != nil {
				return err
			}

			// 3.3 기존 Annotation들의 map을 만듭니다
			existingAnnotationIDs := make(map[int64]bool)
			for _, annotation := range existingAnnotations {
				existingAnnotationIDs[annotation.ID] = true
			}

			// 3.4 새로운 Annotation들을 처리합니다
			for j := range devPlan.Plans[i].Annotations {
				annotation := &devPlan.Plans[i].Annotations[j]
				annotation.PlanID = devPlan.Plans[i].ID

				if annotation.ID > 0 && existingAnnotationIDs[annotation.ID] {
					// 기존 Annotation 업데이트
					if err := s.annotationRepo.UpdateAnnotation(ctx, annotation); err != nil {
						return err
					}
					delete(existingAnnotationIDs, annotation.ID)
				} else {
					// 새로운 Annotation 생성
					annotation.ID = 0 // ID 초기화
					if err := s.annotationRepo.CreateAnnotation(ctx, annotation); err != nil {
						return err
					}
				}
			}

			// 3.5 남은 Annotation들을 삭제합니다
			for annotationID := range existingAnnotationIDs {
				if err := s.annotationRepo.DeleteAnnotation(ctx, annotationID); err != nil {
					return err
				}
			}

			delete(existingPlanIDs, devPlan.Plans[i].ID)
		} else {
			// 3.6 새로운 Plan 생성
			devPlan.Plans[i].ID = 0 // ID 초기화
			if err := s.planRepo.CreatePlan(ctx, &devPlan.Plans[i]); err != nil {
				return err
			}

			// 3.7 새로운 Plan의 Annotation들을 생성합니다
			for j := range devPlan.Plans[i].Annotations {
				annotation := &devPlan.Plans[i].Annotations[j]
				annotation.PlanID = devPlan.Plans[i].ID
				annotation.ID = 0 // ID 초기화
				if err := s.annotationRepo.CreateAnnotation(ctx, annotation); err != nil {
					return err
				}
			}
		}
	}

	// 4. 남은 Plan들과 관련된 모든 Annotation들을 삭제합니다
	for planID := range existingPlanIDs {
		// 4.1 먼저 Plan에 속한 모든 Annotation들을 삭제합니다
		if err := s.annotationRepo.DeleteAnnotationsByPlanID(ctx, planID); err != nil {
			return err
		}
		// 4.2 그 다음 Plan을 삭제합니다
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
