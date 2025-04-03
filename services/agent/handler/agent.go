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

type AgentHandler struct {
	pb.UnimplementedAgentServiceServer
	Config        configs.Config
	VectorDB      VectorDB
	RdbConnection *storage.RDBConnection
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
	planRepo := repo.NewPlanRepository(a.RdbConnection)
	err = planRepo.CreateDevPlanWithDetails(ctx, &model.DevPlan{
		ProjectID: project.ID,
		Branch:    project.Branch,
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
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save dev plan: %v", err)
	}
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

	response := &pb.GeneratePlanResponse{
		Language: devPlan.Language,
		Plans:    plans,
	}
	return response, nil
}

func (a *AgentHandler) ImplementPlan(ctx context.Context, request *pb.ImplementPlanRequest) (*pb.ImplementPlanResponse, error) {
	workerAgent := service.NewWorkerAgent(a.Config.OpenAiKey)
	results, err := workerAgent.ImplementPlan(request.Language, request.Plans)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %v", err)
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
