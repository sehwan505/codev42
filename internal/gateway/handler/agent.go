package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

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

	fmt.Println("ModifyPlan request:", req)

	resp, err := h.grpcClient.ModifyPlan(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AgentHandler) GetPlanList(c *gin.Context) {
	projectID := c.Query("ProjectId")
	branch := c.Query("Branch")

	req := &pb.GetPlanListRequest{
		ProjectId: projectID,
		Branch:    branch,
	}

	resp, err := h.grpcClient.GetPlanList(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AgentHandler) GetPlanById(c *gin.Context) {
	DevPlanID := c.Query("DevPlanId")
	DevPlanIDInt, err := strconv.ParseInt(DevPlanID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := &pb.GetPlanByIdRequest{
		DevPlanId: DevPlanIDInt,
	}

	resp, err := h.grpcClient.GetPlanById(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
