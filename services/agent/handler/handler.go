package handler

import (
	"context"
	"fmt"

	"github.com/sehwan505/codev42/services/agent/configs"
	"github.com/sehwan505/codev42/services/agent/pb"
)

type AgentHandler struct {
	pb.UnimplementedAgentServiceServer
	Config configs.Config
}

func (a *AgentHandler) GeneratePlan(ctx context.Context, request *pb.GeneratePlanRequest) (*pb.GeneratePlanResponse, error) {
	masterAgent := NewMasterAgent(a.Config.OpenAiKey)
	devPlan, err := masterAgent.Call(request.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %v", err)
	}
	response := &pb.GeneratePlanResponse{
		Language: devPlan.Language,
		Plans:    devPlan.Plans,
	}
	return response, nil
}

func (a *AgentHandler) ImplementPlan(ctx context.Context, request *pb.ImplementPlanRequest) (*pb.ImplementPlanResponse, error) {
	workerAgent := NewWorkerAgent(a.Config.OpenAiKey)
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
