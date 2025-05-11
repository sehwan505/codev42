package routes

import (
	"log"

	"codev42-agent/pb"
	"codev42/internal/gateway/handler"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupRoutes() (*grpc.ClientConn, *gin.Engine) {
	router := gin.Default()
	// conn, err := grpc.NewClient("agent-server.default.svc.cluster.local:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}

	agentClient := pb.NewAgentServiceClient(conn)
	agentHandler := handler.NewAgentHandler(agentClient)

	planClient := pb.NewPlanServiceClient(conn)
	planHandler := handler.NewPlanHandler(planClient)

	router.POST("/generate-plan", agentHandler.GeneratePlan)
	router.POST("/implement-plan", agentHandler.ImplementPlan)
	router.POST("/modify-plan", agentHandler.ModifyPlan)
	router.GET("/get-plan-list", planHandler.GetPlanList)
	router.GET("/get-plan-by-id", planHandler.GetPlanById)
	return conn, router
}
