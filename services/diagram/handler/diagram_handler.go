package handler

import (
	"context"
	"fmt"

	"codev42-diagram/configs"
	"codev42-diagram/pb"
	"codev42-diagram/service"
)

type DiagramHandler struct {
	pb.UnimplementedDiagramServiceServer
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

// GenerateDiagrams generates all diagram types in parallel
func (h *DiagramHandler) GenerateDiagrams(ctx context.Context, req *pb.GenerateDiagramsRequest) (*pb.GenerateDiagramsResponse, error) {
	results, err := h.diagramAgent.ImplementDiagrams(req.Code, req.Purpose)
	if err != nil {
		return nil, fmt.Errorf("failed to generate diagrams: %v", err)
	}

	pbResults := make([]*pb.DiagramResult, len(results))
	successCount := 0

	for i, result := range results {
		success := result.Diagram != ""
		if success {
			successCount++
		}

		pbResults[i] = &pb.DiagramResult{
			Diagram: result.Diagram,
			Type:    string(result.Type),
			Success: success,
			Error:   "",
		}
	}

	return &pb.GenerateDiagramsResponse{
		Diagrams:     pbResults,
		SuccessCount: int32(successCount),
		TotalCount:   int32(len(results)),
	}, nil
}

// GenerateClassDiagram generates a class diagram
func (h *DiagramHandler) GenerateClassDiagram(ctx context.Context, req *pb.GenerateDiagramRequest) (*pb.GenerateDiagramResponse, error) {
	result, err := h.diagramAgent.GenerateClassDiagram(req.Code, req.Purpose)
	if err != nil {
		return &pb.GenerateDiagramResponse{
			Diagram: "",
			Type:    "classDiagram",
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GenerateDiagramResponse{
		Diagram: result.Diagram,
		Type:    "classDiagram",
		Success: true,
		Error:   "",
	}, nil
}

// GenerateSequenceDiagram generates a sequence diagram
func (h *DiagramHandler) GenerateSequenceDiagram(ctx context.Context, req *pb.GenerateDiagramRequest) (*pb.GenerateDiagramResponse, error) {
	result, err := h.diagramAgent.GenerateSequenceDiagram(req.Code, req.Purpose)
	if err != nil {
		return &pb.GenerateDiagramResponse{
			Diagram: "",
			Type:    "sequenceDiagram",
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GenerateDiagramResponse{
		Diagram: result.Diagram,
		Type:    "sequenceDiagram",
		Success: true,
		Error:   "",
	}, nil
}

// GenerateFlowchartDiagram generates a flowchart diagram
func (h *DiagramHandler) GenerateFlowchartDiagram(ctx context.Context, req *pb.GenerateDiagramRequest) (*pb.GenerateDiagramResponse, error) {
	result, err := h.diagramAgent.GenerateFlowchartDiagram(req.Code, req.Purpose)
	if err != nil {
		return &pb.GenerateDiagramResponse{
			Diagram: "",
			Type:    "flowchart",
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GenerateDiagramResponse{
		Diagram: result.Diagram,
		Type:    "flowchart",
		Success: true,
		Error:   "",
	}, nil
}
