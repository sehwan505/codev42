package handler

import (
	"context"
	"fmt"

	"codev42-implementation/configs"
	"codev42-implementation/proto/analyzer"
	"codev42-implementation/proto/diagram"
	"codev42-implementation/proto/implementation"
	"codev42-implementation/proto/plan"
	"codev42-implementation/service"
)

type ImplementationHandler struct {
	implementation.UnimplementedImplementationServiceServer
	Config         configs.Config
	workerAgent    *service.WorkerAgent
	planClient     plan.PlanServiceClient
	diagramClient  diagram.DiagramServiceClient
	analyzerClient analyzer.AnalyzerServiceClient
}

func NewImplementationHandler(
	config configs.Config,
	planClient plan.PlanServiceClient,
	diagramClient diagram.DiagramServiceClient,
	analyzerClient analyzer.AnalyzerServiceClient,
) *ImplementationHandler {
	workerAgent := service.NewWorkerAgent(config.OpenAiKey)

	return &ImplementationHandler{
		Config:         config,
		workerAgent:    workerAgent,
		planClient:     planClient,
		diagramClient:  diagramClient,
		analyzerClient: analyzerClient,
	}
}

// ImplementPlan 코드 구현 (동기 실행)
func (h *ImplementationHandler) ImplementPlan(ctx context.Context, req *implementation.ImplementPlanRequest) (*implementation.ImplementPlanResponse, error) {
	// Plan 서비스에서 개발 계획 조회
	planResp, err := h.planClient.GetPlanById(ctx, &plan.GetPlanByIdRequest{
		DevPlanId: req.DevPlanId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plan: %v", err)
	}

	// planpb.Plan -> service.Plan 변환
	plans := make([]service.Plan, 0, len(planResp.Plans))
	for _, pbPlan := range planResp.Plans {
		annotations := make([]service.Annotation, 0, len(pbPlan.Annotations))
		for _, pbAnnotation := range pbPlan.Annotations {
			annotations = append(annotations, service.Annotation{
				Name:        pbAnnotation.Name,
				Description: pbAnnotation.Description,
				Params:      pbAnnotation.Params,
				Returns:     pbAnnotation.Returns,
			})
		}
		plans = append(plans, service.Plan{
			ClassName:   pbPlan.ClassName,
			Annotations: annotations,
		})
	}

	// AI로 코드 생성
	results, err := h.workerAgent.ImplementPlan(planResp.Language, plans)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %v", err)
	}

	// 코드 결과 조합
	var code string
	if len(results) > 0 && results[0] != nil {
		code = results[0].Code
	}

	if code == "" {
		return nil, fmt.Errorf("generated code is empty")
	}

	// Diagram 서비스로 다이어그램 생성
	diagramResp, err := h.diagramClient.GenerateDiagrams(ctx, &diagram.GenerateDiagramsRequest{
		Code:    code,
		Purpose: fmt.Sprintf("Development Plan ID: %d", req.DevPlanId),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate diagrams: %v", err)
	}

	// 다이어그램 결과 변환
	diagrams := make([]*implementation.Diagram, 0, len(diagramResp.Diagrams))
	for _, pbDiagram := range diagramResp.Diagrams {
		diagrams = append(diagrams, &implementation.Diagram{
			Diagram: pbDiagram.Diagram,
			Type:    pbDiagram.Type,
		})
	}

	// 4. Analyzer 서비스로 코드 분석
	analyzerResp, err := h.analyzerClient.AnalyzeCodeSegments(ctx, &analyzer.AnalyzeCodeSegmentsRequest{
		Code:     code,
		Language: planResp.Language,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze code: %v", err)
	}

	// 분석 결과 변환
	explainedSegments := make([]*implementation.ExplainedSegment, 0, len(analyzerResp.CodeSegments))
	for _, pbSegment := range analyzerResp.CodeSegments {
		explainedSegments = append(explainedSegments, &implementation.ExplainedSegment{
			StartLine:   pbSegment.StartLine,
			EndLine:     pbSegment.EndLine,
			Explanation: pbSegment.Explanation,
		})
	}

	// 5. 최종 결과 반환
	return &implementation.ImplementPlanResponse{
		Code:              code,
		Diagrams:          diagrams,
		ExplainedSegments: explainedSegments,
		Status:            "completed",
		Message:           "Implementation completed successfully",
	}, nil
}
