package handler

import (
	"context"
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

	resp, err := h.grpcClient.GeneratePlan(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
