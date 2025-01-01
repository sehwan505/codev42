package grpcserver

import (
	"context"

	"github.com/sehwan505/codev42/pb"
)

type AgentServiceServer struct {
	pb.UnimplementedAgentServiceServer
}

func NewAgentServiceServer() *AgentServiceServer {
	return &AgentServiceServer{}
}

func (s *AgentServiceServer) GeneratePlan(ctx context.Context, req *pb.GeneratePlanRequest) (*pb.GeneratePlanResponse, error) {
	return &pb.GeneratePlanResponse{
		Language: "Python",
		Plans:    []string{"Plan1", "Plan2"},
	}, nil
}

func (s *AgentServiceServer) SearchEntities(ctx context.Context, req *pb.ImplementPlanRequest) (*pb.ImplementPlanResponse, error) {
	return &pb.ImplementPlanResponse{
		Description: []string{"Description1", "Description2"},
		Code:        "Generated code here",
	}, nil
}
