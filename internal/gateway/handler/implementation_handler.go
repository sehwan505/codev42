package handler

import (
	"context"
	"net/http"

	implpb "codev42-implementation/pb"

	"github.com/gin-gonic/gin"
)

type ImplementationHandler struct {
	grpcClient implpb.ImplementationServiceClient
}

func NewImplementationHandler(client implpb.ImplementationServiceClient) *ImplementationHandler {
	return &ImplementationHandler{grpcClient: client}
}

// ImplementPlan 비동기 코드 구현 시작
func (h *ImplementationHandler) ImplementPlan(c *gin.Context) {
	var req implpb.ImplementPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.ImplementPlan(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetImplementationStatus 구현 상태 조회
func (h *ImplementationHandler) GetImplementationStatus(c *gin.Context) {
	var req implpb.GetImplementationStatusRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.GetImplementationStatus(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetImplementationResult 구현 결과 조회
func (h *ImplementationHandler) GetImplementationResult(c *gin.Context) {
	var req implpb.GetImplementationResultRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.GetImplementationResult(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}