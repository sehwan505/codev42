package handler

import (
	"context"
	"fmt"

	"codev42-plan/configs"
	"codev42-plan/model"
	"codev42-plan/proto/plan"
	"codev42-plan/service"
	"codev42-plan/storage"
	"codev42-plan/storage/repo"
)

type PlanHandler struct {
	plan.UnimplementedPlanServiceServer
	Config      configs.Config
	DB          *storage.RDBConnection
	planSvc     *service.PlanService
	masterAgent *service.MasterAgent
}

func NewPlanHandler(config configs.Config, db *storage.RDBConnection) *PlanHandler {
	// 저장소 초기화
	devPlanRepo := repo.NewDevPlanRepository(db)
	planRepo := repo.NewPlanRepository(db)
	annotationRepo := repo.NewAnnotationRepository(db)

	// 서비스 초기화
	planSvc := service.NewPlanService(devPlanRepo, planRepo, annotationRepo)
	masterAgent := service.NewMasterAgent(config.OpenAiKey)

	return &PlanHandler{
		Config:      config,
		DB:          db,
		planSvc:     planSvc,
		masterAgent: masterAgent,
	}
}

// service.DevPlan을 model.DevPlan으로 변환
func convertServiceDevPlanToModelDevPlan(projectID, branch string, devPlan *service.DevPlan, prompt string) *model.DevPlan {
	return &model.DevPlan{
		ProjectID: projectID,
		Branch:    branch,
		Language:  devPlan.Language,
		Prompt:    prompt,
		Plans: func() []model.Plan {
			plans := make([]model.Plan, len(devPlan.Plans))
			for i, plan := range devPlan.Plans {
				plans[i] = model.Plan{
					ClassName: plan.ClassName,
					Annotations: func() []model.Annotation {
						annotations := make([]model.Annotation, len(plan.Annotations))
						for j, ann := range plan.Annotations {
							annotations[j] = model.Annotation{
								Name:        ann.Name,
								Params:      ann.Params,
								Returns:     ann.Returns,
								Description: ann.Description,
							}
						}
						return annotations
					}(),
				}
			}
			return plans
		}(),
	}
}

// model.DevPlan을 plan.GeneratePlanResponse으로 변환
func createPBResponse(devPlan *model.DevPlan) *plan.GeneratePlanResponse {
	pbPlans := make([]*plan.Plan, len(devPlan.Plans))
	for i, modelPlan := range devPlan.Plans {
		pbAnnotations := make([]*plan.Annotation, len(modelPlan.Annotations))
		for j, ann := range modelPlan.Annotations {
			pbAnnotations[j] = &plan.Annotation{
				Name:        ann.Name,
				Params:      ann.Params,
				Returns:     ann.Returns,
				Description: ann.Description,
			}
		}

		pbPlans[i] = &plan.Plan{
			ClassName:   modelPlan.ClassName,
			Annotations: pbAnnotations,
		}
	}

	return &plan.GeneratePlanResponse{
		DevPlanId: devPlan.ID,
		Language:  devPlan.Language,
		Plans:     pbPlans,
	}
}

// 새로운 개발 계획 생성
func (h *PlanHandler) GeneratePlan(ctx context.Context, request *plan.GeneratePlanRequest) (*plan.GeneratePlanResponse, error) {
	// 1. 마스터 에이전트를 사용하여 계획 생성
	devPlan, err := h.masterAgent.Call(request.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %v", err)
	}

	// 2. 프로젝트가 존재하는지 확인, 없으면 생성
	projectRepo := repo.NewProjectRepo(h.DB)
	project, err := projectRepo.GetProjectByID(ctx, request.ProjectId, request.Branch)
	if err != nil {
		// 프로젝트가 존재하지 않으면 생성
		project = &model.Project{
			ID:     request.ProjectId,
			Branch: request.Branch,
			Name:   request.ProjectId,
		}
		if err := projectRepo.CreateProject(ctx, project); err != nil {
			return nil, fmt.Errorf("failed to create project: %v", err)
		}
	}

	// 3. service DevPlan을 model DevPlan으로 변환
	modelDevPlan := convertServiceDevPlanToModelDevPlan(request.ProjectId, request.Branch, devPlan, request.Prompt)

	// 4. 데이터베이스에 저장
	if err := h.planSvc.CreateDevPlanWithDetails(ctx, modelDevPlan); err != nil {
		return nil, fmt.Errorf("failed to save plan: %v", err)
	}

	// 5. 응답 반환
	return createPBResponse(modelDevPlan), nil
}

// 기존 개발 계획 수정
func (h *PlanHandler) ModifyPlan(ctx context.Context, request *plan.ModifyPlanRequest) (*plan.ModifyPlanResponse, error) {
	// plan.Plan을 model.Plan으로 변환
	modelPlans := make([]model.Plan, len(request.Plans))
	for i, pbPlan := range request.Plans {
		annotations := make([]model.Annotation, len(pbPlan.Annotations))
		for j, pbAnn := range pbPlan.Annotations {
			annotations[j] = model.Annotation{
				Name:        pbAnn.Name,
				Params:      pbAnn.Params,
				Returns:     pbAnn.Returns,
				Description: pbAnn.Description,
			}
		}
		modelPlans[i] = model.Plan{
			ClassName:   pbPlan.ClassName,
			Annotations: annotations,
		}
	}

	// 기존 DevPlan 조회
	existingDevPlan, err := h.planSvc.GetDevPlanByID(ctx, request.DevPlanId)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing plan: %v", err)
	}

	// DevPlan 업데이트
	existingDevPlan.Language = request.Language
	existingDevPlan.Plans = modelPlans

	if err := h.planSvc.UpdateDevPlanWithDetails(ctx, existingDevPlan); err != nil {
		return nil, fmt.Errorf("failed to update plan: %v", err)
	}

	return &plan.ModifyPlanResponse{
		Status: "success",
	}, nil
}

// ID로 개발 계획 조회
func (h *PlanHandler) GetPlanById(ctx context.Context, request *plan.GetPlanByIdRequest) (*plan.GetPlanByIdResponse, error) {
	devPlan, err := h.planSvc.GetDevPlanByID(ctx, request.DevPlanId)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %v", err)
	}

	// pb 형식으로 변환
	pbPlans := make([]*plan.Plan, len(devPlan.Plans))
	for i, modelPlan := range devPlan.Plans {
		pbAnnotations := make([]*plan.Annotation, len(modelPlan.Annotations))
		for j, ann := range modelPlan.Annotations {
			pbAnnotations[j] = &plan.Annotation{
				Name:        ann.Name,
				Params:      ann.Params,
				Returns:     ann.Returns,
				Description: ann.Description,
			}
		}
		pbPlans[i] = &plan.Plan{
			ClassName:   modelPlan.ClassName,
			Annotations: pbAnnotations,
		}
	}

	return &plan.GetPlanByIdResponse{
		DevPlanId: devPlan.ID,
		ProjectId: devPlan.ProjectID,
		Branch:    devPlan.Branch,
		Language:  devPlan.Language,
		Plans:     pbPlans,
	}, nil
}

// 프로젝트의 모든 개발 계획 조회
func (h *PlanHandler) GetPlanList(ctx context.Context, request *plan.GetPlanListRequest) (*plan.GetPlanListResponse, error) {
	devPlans, err := h.planSvc.GetDevPlansByProjectID(ctx, request.ProjectId, request.Branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan list: %v", err)
	}

	// pb 형식으로 변환
	pbList := make([]*plan.PlanListElement, len(devPlans))
	for i, dp := range devPlans {
		pbList[i] = &plan.PlanListElement{
			DevPlanId: dp.ID,
			Prompt:    dp.Prompt,
		}
	}

	return &plan.GetPlanListResponse{
		DevPlanList: pbList,
	}, nil
}
