package handler

import (
	"codev42-agent/pb"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PlanHandler struct {
	grpcClient pb.PlanServiceClient
}

func NewPlanHandler(client pb.PlanServiceClient) *PlanHandler {
	return &PlanHandler{grpcClient: client}
}

func (h *PlanHandler) GetPlanList(c *gin.Context) {
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

func (h *PlanHandler) GetPlanById(c *gin.Context) {
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
