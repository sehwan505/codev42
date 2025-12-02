package handler

import (
	"context"
	"fmt"

	"codev42-diagram/configs"
	"codev42-diagram/proto/diagram"
	"codev42-diagram/service"
)

type DiagramHandler struct {
	diagram.UnimplementedDiagramServiceServer
	Config       configs.Config
	diagramAgent *service.DiagramAgent
}

func NewDiagramHandler(config configs.Config) *DiagramHandler {
	diagramAgent := service.NewDiagramAgent(config.OpenAiKey)

	return &DiagramHandler{
		Config:       config,
		diagramAgent: diagramAgent,
	}
}

// GenerateDiagrams 모든 다이어그램 병렬 생성
func (h *DiagramHandler) GenerateDiagrams(ctx context.Context, req *diagram.GenerateDiagramsRequest) (*diagram.GenerateDiagramsResponse, error) {
	results, err := h.diagramAgent.ImplementDiagrams(req.Code, req.Purpose)
	if err != nil {
		return nil, fmt.Errorf("failed to generate diagrams: %v", err)
	}

	pbResults := make([]*diagram.DiagramResult, len(results))
	successCount := 0

	for i, result := range results {
		success := result.Diagram != ""
		if success {
			successCount++
		}

		pbResults[i] = &diagram.DiagramResult{
			Diagram: result.Diagram,
			Type:    string(result.Type),
			Success: success,
			Error:   "",
		}
	}

	return &diagram.GenerateDiagramsResponse{
		Diagrams:     pbResults,
		SuccessCount: int32(successCount),
		TotalCount:   int32(len(results)),
	}, nil
}

// GenerateClassDiagram 클래스 다이어그램 생성
func (h *DiagramHandler) GenerateClassDiagram(ctx context.Context, req *diagram.GenerateDiagramRequest) (*diagram.GenerateDiagramResponse, error) {
	result, err := h.diagramAgent.GenerateClassDiagram(req.Code, req.Purpose)
	if err != nil {
		return &diagram.GenerateDiagramResponse{
			Diagram: "",
			Type:    "classDiagram",
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &diagram.GenerateDiagramResponse{
		Diagram: result.Diagram,
		Type:    "classDiagram",
		Success: true,
		Error:   "",
	}, nil
}

// GenerateSequenceDiagram 시퀀스 다이어그램 생성
func (h *DiagramHandler) GenerateSequenceDiagram(ctx context.Context, req *diagram.GenerateDiagramRequest) (*diagram.GenerateDiagramResponse, error) {
	result, err := h.diagramAgent.GenerateSequenceDiagram(req.Code, req.Purpose)
	if err != nil {
		return &diagram.GenerateDiagramResponse{
			Diagram: "",
			Type:    "sequenceDiagram",
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &diagram.GenerateDiagramResponse{
		Diagram: result.Diagram,
		Type:    "sequenceDiagram",
		Success: true,
		Error:   "",
	}, nil
}

// GenerateFlowchartDiagram 플로우차트 생성
func (h *DiagramHandler) GenerateFlowchartDiagram(ctx context.Context, req *diagram.GenerateDiagramRequest) (*diagram.GenerateDiagramResponse, error) {
	result, err := h.diagramAgent.GenerateFlowchartDiagram(req.Code, req.Purpose)
	if err != nil {
		return &diagram.GenerateDiagramResponse{
			Diagram: "",
			Type:    "flowchart",
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &diagram.GenerateDiagramResponse{
		Diagram: result.Diagram,
		Type:    "flowchart",
		Success: true,
		Error:   "",
	}, nil
}
