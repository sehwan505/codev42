package handler

import (
	"context"
	"fmt"

	"codev42-analyzer/configs"
	"codev42-analyzer/proto/analyzer"
	"codev42-analyzer/service"
)

type AnalyzerHandler struct {
	analyzer.UnimplementedAnalyzerServiceServer
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

// CombineCode 여러 코드 조각을 하나로 조합
func (h *AnalyzerHandler) CombineCode(ctx context.Context, req *analyzer.CombineCodeRequest) (*analyzer.CombineCodeResponse, error) {
	var implementResults []*service.ImplementResult
	for _, code := range req.Codes {
		implementResults = append(implementResults, &service.ImplementResult{
			Code: code,
		})
	}

	result, err := h.analyserAgent.CombineImplementation(implementResults, req.Purpose)
	if err != nil {
		return &analyzer.CombineCodeResponse{
			Code:    "",
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &analyzer.CombineCodeResponse{
		Code:    result.Code,
		Success: true,
		Error:   "",
	}, nil
}

// AnalyzeCodeSegments 코드 분석 및 설명 생성
func (h *AnalyzerHandler) AnalyzeCodeSegments(ctx context.Context, req *analyzer.AnalyzeCodeSegmentsRequest) (*analyzer.AnalyzeCodeSegmentsResponse, error) {
	segments, err := h.analyserAgent.AnalyzeCodeSegments(req.Code, req.Language)
	if err != nil {
		return &analyzer.AnalyzeCodeSegmentsResponse{
			CodeSegments: nil,
			Success:      false,
			Error:        fmt.Sprintf("failed to analyze code segments: %v", err),
		}, nil
	}
	pbSegments := make([]*analyzer.CodeSegment, len(segments))
	for i, segment := range segments {
		pbSegments[i] = &analyzer.CodeSegment{
			StartLine:   int32(segment.StartLine),
			EndLine:     int32(segment.EndLine),
			Explanation: segment.Explanation,
		}
	}

	return &analyzer.AnalyzeCodeSegmentsResponse{
		CodeSegments: pbSegments,
		Success:      true,
		Error:        "",
	}, nil
}
