package handler

import (
	"context"
	"fmt"

	"codev42-plan/configs"
	"codev42-plan/model"
	"codev42-plan/pb"
	"codev42-plan/service"
	"codev42-plan/storage"
	"codev42-plan/storage/repo"
)

type PlanHandler struct {
	pb.UnimplementedPlanServiceServer
	Config     configs.Config
	DB         *storage.RDBConnection
	planSvc    *service.PlanService
	masterAgent *service.MasterAgent
}

func NewPlanHandler(config configs.Config, db *storage.RDBConnection) *PlanHandler {
	// Initialize repositories
	devPlanRepo := repo.NewDevPlanRepository(db)
	planRepo := repo.NewPlanRepository(db)
	annotationRepo := repo.NewAnnotationRepository(db)

	// Initialize services
	planSvc := service.NewPlanService(devPlanRepo, planRepo, annotationRepo)
	masterAgent := service.NewMasterAgent(config.OpenAiKey)

	return &PlanHandler{
		Config:      config,
		DB:          db,
		planSvc:     planSvc,
		masterAgent: masterAgent,
	}
}

// Helper: Convert service.DevPlan to model.DevPlan
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

// Helper: Convert model.DevPlan to pb.GeneratePlanResponse
func createPBResponse(devPlan *model.DevPlan) *pb.GeneratePlanResponse {
	plans := make([]*pb.Plan, len(devPlan.Plans))
	for i, plan := range devPlan.Plans {
		annotations := make([]*pb.Annotation, len(plan.Annotations))
		for j, ann := range plan.Annotations {
			annotations[j] = &pb.Annotation{
				Name:        ann.Name,
				Params:      ann.Params,
				Returns:     ann.Returns,
				Description: ann.Description,
			}
		}

		plans[i] = &pb.Plan{
			ClassName:   plan.ClassName,
			Annotations: annotations,
		}
	}

	return &pb.GeneratePlanResponse{
		DevPlanId: devPlan.ID,
		Language:  devPlan.Language,
		Plans:     plans,
	}
}

// GeneratePlan creates a new development plan
func (h *PlanHandler) GeneratePlan(ctx context.Context, request *pb.GeneratePlanRequest) (*pb.GeneratePlanResponse, error) {
	// 1. Generate plan using Master Agent
	devPlan, err := h.masterAgent.Call(request.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %v", err)
	}

	// 2. Check if project exists, create if not
	projectRepo := repo.NewProjectRepo(h.DB)
	project, err := projectRepo.GetProjectByID(ctx, request.ProjectId, request.Branch)
	if err != nil {
		// Project doesn't exist, create it
		project = &model.Project{
			ID:     request.ProjectId,
			Branch: request.Branch,
			Name:   request.ProjectId,
		}
		if err := projectRepo.CreateProject(ctx, project); err != nil {
			return nil, fmt.Errorf("failed to create project: %v", err)
		}
	}

	// 3. Convert service DevPlan to model DevPlan
	modelDevPlan := convertServiceDevPlanToModelDevPlan(request.ProjectId, request.Branch, devPlan, request.Prompt)

	// 4. Save to database
	if err := h.planSvc.CreateDevPlanWithDetails(ctx, modelDevPlan); err != nil {
		return nil, fmt.Errorf("failed to save plan: %v", err)
	}

	// 5. Return response
	return createPBResponse(modelDevPlan), nil
}

// ModifyPlan updates an existing development plan
func (h *PlanHandler) ModifyPlan(ctx context.Context, request *pb.ModifyPlanRequest) (*pb.ModifyPlanResponse, error) {
	// 1. Convert pb.Plan to model.Plan
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

	// 2. Get existing DevPlan
	existingDevPlan, err := h.planSvc.GetDevPlanByID(ctx, request.DevPlanId)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing plan: %v", err)
	}

	// 3. Update DevPlan
	existingDevPlan.Language = request.Language
	existingDevPlan.Plans = modelPlans

	if err := h.planSvc.UpdateDevPlanWithDetails(ctx, existingDevPlan); err != nil {
		return nil, fmt.Errorf("failed to update plan: %v", err)
	}

	return &pb.ModifyPlanResponse{
		Status: "success",
	}, nil
}

// GetPlanById retrieves a development plan by ID
func (h *PlanHandler) GetPlanById(ctx context.Context, request *pb.GetPlanByIdRequest) (*pb.GetPlanByIdResponse, error) {
	devPlan, err := h.planSvc.GetDevPlanByID(ctx, request.DevPlanId)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %v", err)
	}

	// Convert to pb response
	pbPlans := make([]*pb.Plan, len(devPlan.Plans))
	for i, plan := range devPlan.Plans {
		pbAnnotations := make([]*pb.Annotation, len(plan.Annotations))
		for j, ann := range plan.Annotations {
			pbAnnotations[j] = &pb.Annotation{
				Name:        ann.Name,
				Params:      ann.Params,
				Returns:     ann.Returns,
				Description: ann.Description,
			}
		}
		pbPlans[i] = &pb.Plan{
			ClassName:   plan.ClassName,
			Annotations: pbAnnotations,
		}
	}

	return &pb.GetPlanByIdResponse{
		DevPlanId: devPlan.ID,
		ProjectId: devPlan.ProjectID,
		Branch:    devPlan.Branch,
		Language:  devPlan.Language,
		Plans:     pbPlans,
	}, nil
}

// GetPlanList retrieves all development plans for a project
func (h *PlanHandler) GetPlanList(ctx context.Context, request *pb.GetPlanListRequest) (*pb.GetPlanListResponse, error) {
	devPlans, err := h.planSvc.GetDevPlansByProjectID(ctx, request.ProjectId, request.Branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan list: %v", err)
	}

	// Convert to pb response
	pbList := make([]*pb.PlanListElement, len(devPlans))
	for i, dp := range devPlans {
		pbList[i] = &pb.PlanListElement{
			DevPlanId: dp.ID,
			Prompt:    dp.Prompt,
		}
	}

	return &pb.GetPlanListResponse{
		DevPlanList: pbList,
	}, nil
}
