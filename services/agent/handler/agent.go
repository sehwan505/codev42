package handler

import (
	"context"
	"fmt"

	"codev42-agent/configs"
	"codev42-agent/pb"
	"codev42-agent/service"
)

type AgentHandler struct {
	pb.UnimplementedAgentServiceServer
	Config configs.Config
}

func (a *AgentHandler) GeneratePlan(ctx context.Context, request *pb.GeneratePlanRequest) (*pb.GeneratePlanResponse, error) {
	masterAgent := service.NewMasterAgent(a.Config.OpenAiKey)
	devPlan, err := masterAgent.Call(request.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %v", err)
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
