package handler

import (
	"context"
	"fmt"
	"net/http"

	pb "codev42-agent/pb"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	grpcClient pb.AgentServiceClient
}

func NewAgentHandler(client pb.AgentServiceClient) *AgentHandler {
	return &AgentHandler{grpcClient: client}
}

func (h *AgentHandler) GeneratePlan(c *gin.Context) {
	var req pb.GeneratePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 파싱된 데이터 로깅
	fmt.Printf("Parsed request: Prompt=%s, ProjectId=%s, Branch=%s\n",
		req.Prompt, req.ProjectId, req.Branch)

	resp, err := h.grpcClient.GeneratePlan(context.Background(), &req)
	if err != nil {
		fmt.Printf("gRPC 호출 오류: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("GeneratePlan response:", resp)

	c.JSON(http.StatusOK, resp)
}

func (h *AgentHandler) ImplementPlan(c *gin.Context) {
	var req pb.ImplementPlanRequest
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

func (h *AgentHandler) ModifyPlan(c *gin.Context) {
	var req pb.ModifyPlanRequest
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
