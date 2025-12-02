package handler

import (
	"context"
	"fmt"
	"net/http"

	planpb "codev42-plan/proto/plan"

	"github.com/gin-gonic/gin"
)

type PlanHandler struct {
	grpcClient planpb.PlanServiceClient
}

func NewPlanHandler(client planpb.PlanServiceClient) *PlanHandler {
	return &PlanHandler{grpcClient: client}
}

// GeneratePlan 개발 계획 생성
func (h *PlanHandler) GeneratePlan(c *gin.Context) {
	var req planpb.GeneratePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.GeneratePlan(context.Background(), &req)
	if err != nil {
		fmt.Printf("gRPC 호출 오류: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ModifyPlan 개발 계획 수정
func (h *PlanHandler) ModifyPlan(c *gin.Context) {
	var req planpb.ModifyPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.ModifyPlan(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetPlanList 계획 목록 조회
func (h *PlanHandler) GetPlanList(c *gin.Context) {
	var req planpb.GetPlanListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.GetPlanList(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetPlanById ID로 계획 조회
func (h *PlanHandler) GetPlanById(c *gin.Context) {
	var req planpb.GetPlanByIdRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.GetPlanById(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}