package handler

import (
	"context"
	"net/http"

	diagrampb "codev42-diagram/pb"

	"github.com/gin-gonic/gin"
)

type DiagramHandler struct {
	grpcClient diagrampb.DiagramServiceClient
}

func NewDiagramHandler(client diagrampb.DiagramServiceClient) *DiagramHandler {
	return &DiagramHandler{grpcClient: client}
}

// GenerateDiagrams 모든 다이어그램 병렬 생성
func (h *DiagramHandler) GenerateDiagrams(c *gin.Context) {
	var req diagrampb.GenerateDiagramsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.GenerateDiagrams(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GenerateClassDiagram 클래스 다이어그램 생성
func (h *DiagramHandler) GenerateClassDiagram(c *gin.Context) {
	var req diagrampb.GenerateDiagramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.GenerateClassDiagram(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GenerateSequenceDiagram 시퀀스 다이어그램 생성
func (h *DiagramHandler) GenerateSequenceDiagram(c *gin.Context) {
	var req diagrampb.GenerateDiagramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.GenerateSequenceDiagram(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GenerateFlowchartDiagram 플로우차트 생성
func (h *DiagramHandler) GenerateFlowchartDiagram(c *gin.Context) {
	var req diagrampb.GenerateDiagramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.grpcClient.GenerateFlowchartDiagram(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}