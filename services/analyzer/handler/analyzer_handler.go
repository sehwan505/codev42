package handler

import (
	"context"
	"fmt"

	"codev42-analyzer/configs"
	"codev42-analyzer/pb"
	"codev42-analyzer/service"
)

type AnalyzerHandler struct {
	pb.UnimplementedAnalyzerServiceServer
	Config        configs.Config
	analyserAgent *service.AnalyserAgent
}

func NewAnalyzerHandler(config configs.Config) *AnalyzerHandler {
	analyserAgent := service.NewAnalyserAgent(config.OpenAiKey)

	return &AnalyzerHandler{
		Config:        config,
		analyserAgent: analyserAgent,
	}
}

// CombineCode combines multiple code snippets into one
func (h *AnalyzerHandler) CombineCode(ctx context.Context, req *pb.CombineCodeRequest) (*pb.CombineCodeResponse, error) {
	// Convert string codes to ImplementResult format
	var implementResults []*service.ImplementResult
	for _, code := range req.Codes {
		implementResults = append(implementResults, &service.ImplementResult{
			Code: code,
		})
	}

	result, err := h.analyserAgent.CombineImplementation(implementResults, req.Purpose)
	if err != nil {
		return &pb.CombineCodeResponse{
			Code:    "",
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.CombineCodeResponse{
		Code:    result.Code,
		Success: true,
		Error:   "",
	}, nil
}

// AnalyzeCodeSegments analyzes code and returns important segments with explanations
func (h *AnalyzerHandler) AnalyzeCodeSegments(ctx context.Context, req *pb.AnalyzeCodeSegmentsRequest) (*pb.AnalyzeCodeSegmentsResponse, error) {
	segments, err := h.analyserAgent.AnalyzeCodeSegments(req.Code, req.Language)
	if err != nil {
		return &pb.AnalyzeCodeSegmentsResponse{
			CodeSegments: nil,
			Success:      false,
			Error:        fmt.Sprintf("failed to analyze code segments: %v", err),
		}, nil
	}

	// Convert to protobuf format
	pbSegments := make([]*pb.CodeSegment, len(segments))
	for i, segment := range segments {
		pbSegments[i] = &pb.CodeSegment{
			StartLine:   int32(segment.StartLine),
			EndLine:     int32(segment.EndLine),
			Explanation: segment.Explanation,
		}
	}

	return &pb.AnalyzeCodeSegmentsResponse{
		CodeSegments: pbSegments,
		Success:      true,
		Error:        "",
	}, nil
}
