package handler

import (
	"context"
	"fmt"
	"strings"

	"codev42-agent/configs"
	"codev42-agent/model"
	"codev42-agent/pb"
	"codev42-agent/service"
	"codev42-agent/storage"
	"codev42-agent/storage/repo"
)

type VectorDB interface {
	InitCollection(ctx context.Context, collectionName string, vectorDim int32) error
	InsertEmbedding(ctx context.Context, collectionName string, id string, embedding []float32) error
	SearchByVector(ctx context.Context, collectionName string, searchVector []float32, topK int) ([]int64, error)
	DeleteByID(ctx context.Context, collectionName string, id string) error
	Close() error
}
type AgentHandler struct {
	pb.UnimplementedAgentServiceServer
	Config        configs.Config
	VectorDB      VectorDB
	RdbConnection *storage.RDBConnection
}

// createPlanService creates and returns a new PlanService instance
func (a *AgentHandler) createPlanService() *service.PlanService {
	devPlanRepo := repo.NewDevPlanRepository(a.RdbConnection)
	planRepo := repo.NewPlanRepository(a.RdbConnection)
	annotationRepo := repo.NewAnnotationRepository(a.RdbConnection)
	return service.NewPlanService(devPlanRepo, planRepo, annotationRepo)
}

func convertServiceDevPlanToModelDevPlan(projectID, branch string, devPlan *service.DevPlan) *model.DevPlan {
	return &model.DevPlan{
		ProjectID: projectID,
		Branch:    branch,
		Language:  devPlan.Language,
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

func createPBResponse(devPlan *service.DevPlan) *pb.GeneratePlanResponse {
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
		Language: devPlan.Language,
		Plans:    plans,
	}
}

func (a *AgentHandler) GeneratePlan(ctx context.Context, request *pb.GeneratePlanRequest) (*pb.GeneratePlanResponse, error) {
	masterAgent := service.NewMasterAgent(a.Config.OpenAiKey)
	devPlan, err := masterAgent.Call(request.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %v", err)
	}

	projectRepo := repo.NewProjectRepo(a.RdbConnection)
	project, err := projectRepo.GetProjectByID(ctx, request.ProjectId, request.Branch)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			splitedProjectId := strings.Split(request.ProjectId, "/")
			project = &model.Project{
				ID:          request.ProjectId,
				Branch:      request.Branch,
				Name:        splitedProjectId[len(splitedProjectId)-1],
				Description: fmt.Sprintf("%s project", splitedProjectId[len(splitedProjectId)-1]),
			}
			if err = projectRepo.CreateProject(ctx, project); err != nil {
				return nil, fmt.Errorf("project: %w", err)
			}
		} else {
			return nil, fmt.Errorf("project fail to find: %w", err)
		}
	}
	planService := a.createPlanService()
	modelDevPlan := convertServiceDevPlanToModelDevPlan(project.ID, project.Branch, devPlan)
	err = planService.CreateDevPlanWithDetails(ctx, modelDevPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to save dev plan: %v", err)
	}

	return createPBResponse(devPlan), nil
}

func (a *AgentHandler) UpdatePlan(ctx context.Context, request *pb.UpdatePlanRequest) (*pb.UpdatePlanResponse, error) {
	planService := a.createPlanService()

	existingPlan, err := planService.GetDevPlanByID(ctx, request.PlanId)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing dev plan: %v", err)
	}

	updatedPlanData := &model.DevPlan{
		ID:        request.PlanId,
		ProjectID: existingPlan.ProjectID,
		Branch:    existingPlan.Branch,
		Language:  request.Language,
		Plans:     make([]model.Plan, len(request.Plans)),
	}

	// 3. Update plan data preserving IDs when appropriate
	for i, plan := range request.Plans {
		modelPlan := model.Plan{
			ClassName:   plan.ClassName,
			Annotations: make([]model.Annotation, len(plan.Annotations)),
		}

		// If we have a matching index in the existing plan, preserve the ID
		if i < len(existingPlan.Plans) {
			modelPlan.ID = existingPlan.Plans[i].ID
		}

		for j, ann := range plan.Annotations {
			modelAnnotation := model.Annotation{
				Name:        ann.Name,
				Params:      ann.Params,
				Returns:     ann.Returns,
				Description: ann.Description,
			}

			// If we have a matching annotation index, preserve the ID
			if i < len(existingPlan.Plans) && j < len(existingPlan.Plans[i].Annotations) {
				modelAnnotation.ID = existingPlan.Plans[i].Annotations[j].ID
			}

			modelPlan.Annotations[j] = modelAnnotation
		}
		updatedPlanData.Plans[i] = modelPlan
	}

	// 4. Update the plan
	err = planService.UpdateDevPlanWithDetails(ctx, updatedPlanData)
	if err != nil {
		return nil, fmt.Errorf("failed to update dev plan: %v", err)
	}

	return &pb.UpdatePlanResponse{
		Success: true,
	}, nil
}

func (a *AgentHandler) GetPlanById(ctx context.Context, request *pb.GetPlanByIdRequest) (*pb.GetPlanByIdResponse, error) {
	planService := a.createPlanService()

	devPlan, err := planService.GetDevPlanByID(ctx, request.PlanId)
	if err != nil {
		return nil, fmt.Errorf("failed to get dev plan: %v", err)
	}

	// Convert to response
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

	return &pb.GetPlanByIdResponse{
		PlanId:    devPlan.ID,
		ProjectId: devPlan.ProjectID,
		Branch:    devPlan.Branch,
		Language:  devPlan.Language,
		Plans:     plans,
	}, nil
}

func (a *AgentHandler) ImplementPlan(ctx context.Context, request *pb.ImplementPlanRequest) (*pb.ImplementPlanResponse, error) {
	workerAgent := service.NewWorkerAgent(a.Config.OpenAiKey)
	results, err := workerAgent.ImplementPlan(request.Language, request.Plans)
	if err != nil {
		return nil, fmt.Errorf("failed to implement plan: %v", err)
	}
	var pbResults []*pb.DevResult
	for _, result := range results {
		pbResult := &pb.DevResult{
			Description: result.Description,
			Code:        result.Code,
		}
		pbResults = append(pbResults, pbResult)
	}
	response := &pb.ImplementPlanResponse{
		DevResults: pbResults,
	}
	return response, nil
}
