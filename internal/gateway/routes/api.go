package routes

import (
	"log"

	analyzerpb "codev42-analyzer/pb"
	diagrampb "codev42-diagram/pb"
	implpb "codev42-implementation/pb"
	planpb "codev42-plan/pb"
	"codev42/internal/gateway/handler"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupRoutes() ([]*grpc.ClientConn, *gin.Engine) {
	router := gin.Default()

	planConn, err := grpc.NewClient("localhost:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Plan service: %v", err)
	}

	implConn, err := grpc.NewClient("localhost:9092", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Implementation service: %v", err)
	}

	diagramConn, err := grpc.NewClient("localhost:9093", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Diagram service: %v", err)
	}

	analyzerConn, err := grpc.NewClient("localhost:9094", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Analyzer service: %v", err)
	}

	planClient := planpb.NewPlanServiceClient(planConn)
	planHandler := handler.NewPlanHandler(planClient)

	implClient := implpb.NewImplementationServiceClient(implConn)
	implHandler := handler.NewImplementationHandler(implClient)

	diagramClient := diagrampb.NewDiagramServiceClient(diagramConn)
	diagramHandler := handler.NewDiagramHandler(diagramClient)

	analyzerClient := analyzerpb.NewAnalyzerServiceClient(analyzerConn)
	analyzerHandler := handler.NewAnalyzerHandler(analyzerClient)

	// Plan endpoints
	router.POST("/generate-plan", planHandler.GeneratePlan)
	router.POST("/modify-plan", planHandler.ModifyPlan)
	router.GET("/get-plan-list", planHandler.GetPlanList)
	router.GET("/get-plan-by-id", planHandler.GetPlanById)

	// Implementation endpoints
	router.POST("/implement-plan", implHandler.ImplementPlan)
	router.GET("/implementation-status", implHandler.GetImplementationStatus)
	router.GET("/implementation-result", implHandler.GetImplementationResult)

	// Diagram endpoints
	router.POST("/generate-diagrams", diagramHandler.GenerateDiagrams)
	router.POST("/generate-class-diagram", diagramHandler.GenerateClassDiagram)
	router.POST("/generate-sequence-diagram", diagramHandler.GenerateSequenceDiagram)
	router.POST("/generate-flowchart-diagram", diagramHandler.GenerateFlowchartDiagram)

	// Analyzer endpoints
	router.POST("/combine-code", analyzerHandler.CombineCode)
	router.POST("/analyze-code-segments", analyzerHandler.AnalyzeCodeSegments)

	return []*grpc.ClientConn{planConn, implConn, diagramConn, analyzerConn}, router
}