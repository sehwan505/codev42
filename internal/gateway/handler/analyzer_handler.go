package handler

import (
	"context"
	"net/http"

	analyzerpb "codev42-analyzer/pb"

	"github.com/gin-gonic/gin"
)

type AnalyzerHandler struct {
	grpcClient analyzerpb.AnalyzerServiceClient
}

func NewAnalyzerHandler(client analyzerpb.AnalyzerServiceClient) *AnalyzerHandler {
	return &AnalyzerHandler{grpcClient: client}
}

// CombineCode 여러 코드 조각을 하나로 조합
func (h *AnalyzerHandler) CombineCode(c *gin.Context) {
	var req analyzerpb.CombineCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.CombineCode(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AnalyzeCodeSegments 코드 분석 및 설명 생성
func (h *AnalyzerHandler) AnalyzeCodeSegments(c *gin.Context) {
	var req analyzerpb.AnalyzeCodeSegmentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.AnalyzeCodeSegments(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}