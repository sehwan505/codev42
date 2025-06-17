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

func (a *AgentHandler) createPlanService() *service.PlanService {
	devPlanRepo := repo.NewDevPlanRepository(a.RdbConnection)
	planRepo := repo.NewPlanRepository(a.RdbConnection)
	annotationRepo := repo.NewAnnotationRepository(a.RdbConnection)
	return service.NewPlanService(devPlanRepo, planRepo, annotationRepo)
}

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
	modelDevPlan := convertServiceDevPlanToModelDevPlan(project.ID, project.Branch, devPlan, request.Prompt)
	err = planService.CreateDevPlanWithDetails(ctx, modelDevPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to save dev plan: %v", err)
	}

	return createPBResponse(modelDevPlan), nil
}

func (a *AgentHandler) ModifyPlan(ctx context.Context, request *pb.ModifyPlanRequest) (*pb.ModifyPlanResponse, error) {
	planService := a.createPlanService()

	existingPlan, err := planService.GetDevPlanByID(ctx, request.DevPlanId)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing dev plan: %v", err)
	}

	updatedPlanData := &model.DevPlan{
		ID:        request.DevPlanId,
		ProjectID: existingPlan.ProjectID,
		Branch:    existingPlan.Branch,
		Language:  request.Language,
		Prompt:    existingPlan.Prompt,
		Plans:     make([]model.Plan, len(request.Plans)),
	}

	for i, plan := range request.Plans {
		modelPlan := model.Plan{
			ClassName:   plan.ClassName,
			Annotations: make([]model.Annotation, len(plan.Annotations)),
		}

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

			if i < len(existingPlan.Plans) && j < len(existingPlan.Plans[i].Annotations) {
				modelAnnotation.ID = existingPlan.Plans[i].Annotations[j].ID
			}

			modelPlan.Annotations[j] = modelAnnotation
		}
		updatedPlanData.Plans[i] = modelPlan
	}

	err = planService.UpdateDevPlanWithDetails(ctx, updatedPlanData)
	if err != nil {
		return nil, fmt.Errorf("failed to update dev plan: %v", err)
	}

	return &pb.ModifyPlanResponse{Status: "success"}, nil
}

func (a *AgentHandler) ImplementPlan(ctx context.Context, request *pb.ImplementPlanRequest) (*pb.ImplementPlanResponse, error) {
	workerAgent := service.NewWorkerAgent(a.Config.OpenAiKey)
	analyserAgent := service.NewAnalyserAgent(a.Config.OpenAiKey)
	diagramAgent := service.NewDiagramAgent(a.Config.OpenAiKey)
	planService := a.createPlanService()
	existingPlan, err := planService.GetDevPlanByID(ctx, request.DevPlanId)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing dev plan: %v", err)
	}
	results, err := workerAgent.ImplementPlan(existingPlan.Language, existingPlan.Plans)
	if err != nil {
		return nil, fmt.Errorf("failed to implement plan: %v", err)
	}
	code := ""
	for _, result := range results {
		code += result.Code + "\n"
	}
	combinedResult, err := analyserAgent.CombineImplementation(results, existingPlan.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to combine implementation: %v", err)
	}
	explainedSegments, err := analyserAgent.AnalyzeCodeSegments(combinedResult.Code, existingPlan.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze code segments: %v", err)
	}
	segmentsPB := make([]*pb.ExplainedSegment, len(explainedSegments))
	for i, segment := range explainedSegments {
		segmentsPB[i] = &pb.ExplainedSegment{
			StartLine:   int32(segment.StartLine),
			EndLine:     int32(segment.EndLine),
			Explanation: segment.Explanation,
		}
	}
	diagrams, err := diagramAgent.ImplementDiagrams(combinedResult.Code, existingPlan.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to implement diagram: %v", err)
	}
	diagramsPB := make([]*pb.Diagram, len(diagrams))
	for i, diagram := range diagrams {
		diagramsPB[i] = &pb.Diagram{
			Diagram: diagram.Diagram,
			Type:    string(diagram.Type),
		}
	}
	return &pb.ImplementPlanResponse{
		Code:              combinedResult.Code,
		Diagrams:          diagramsPB,
		ExplainedSegments: segmentsPB,
	}, nil
}
