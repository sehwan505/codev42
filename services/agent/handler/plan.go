package handler

import (
	"codev42-agent/configs"
	"codev42-agent/pb"
	"codev42-agent/service"
	"codev42-agent/storage"
	"codev42-agent/storage/repo"
	"context"
	"fmt"
)

type PlanHandler struct {
	pb.UnimplementedPlanServiceServer
	Config        configs.Config
	RdbConnection *storage.RDBConnection
}

func (p *PlanHandler) createPlanService() *service.PlanService {
	devPlanRepo := repo.NewDevPlanRepository(p.RdbConnection)
	planRepo := repo.NewPlanRepository(p.RdbConnection)
	annotationRepo := repo.NewAnnotationRepository(p.RdbConnection)
	return service.NewPlanService(devPlanRepo, planRepo, annotationRepo)
}

func (p *PlanHandler) GetPlanList(ctx context.Context, request *pb.GetPlanListRequest) (*pb.GetPlanListResponse, error) {
	planService := p.createPlanService()
	devPlans, err := planService.GetDevPlansByProjectID(ctx, request.ProjectId, request.Branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get dev plan list: %v", err)
	}

	// DevPlan 목록을 PB 형식으로 변환
	var pbDevPlans []*pb.PlanListElement
	for _, plan := range devPlans {
		pbDevPlans = append(pbDevPlans, &pb.PlanListElement{
			DevPlanId: plan.ID,
			Prompt:    plan.Prompt,
		})
	}

	return &pb.GetPlanListResponse{
		DevPlanList: pbDevPlans,
	}, nil
}

func (p *PlanHandler) GetPlanById(ctx context.Context, request *pb.GetPlanByIdRequest) (*pb.GetPlanByIdResponse, error) {
	planService := p.createPlanService()

	devPlan, err := planService.GetDevPlanByID(ctx, request.DevPlanId)
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
		DevPlanId: devPlan.ID,
		ProjectId: devPlan.ProjectID,
		Branch:    devPlan.Branch,
		Language:  devPlan.Language,
		Plans:     plans,
	}, nil
}
