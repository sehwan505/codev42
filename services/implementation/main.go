package main

import (
	"fmt"
	"log"
	"net"

	"codev42-implementation/configs"
	"codev42-implementation/handler"
	"codev42-implementation/proto/analyzer"
	"codev42-implementation/proto/diagram"
	"codev42-implementation/proto/implementation"
	"codev42-implementation/proto/plan"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := configs.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Implementation Service configuration loaded")

	// 다른 서비스로의 gRPC 클라이언트 연결 생성
	log.Printf("Connecting to Plan Service at %s", config.PlanServiceAddr)
	planConn, err := grpc.NewClient(
		config.PlanServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Plan Service: %v", err)
	}
	defer planConn.Close()
	planClient := plan.NewPlanServiceClient(planConn)

	log.Printf("Connecting to Diagram Service at %s", config.DiagramServiceAddr)
	diagramConn, err := grpc.NewClient(
		config.DiagramServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Diagram Service: %v", err)
	}
	defer diagramConn.Close()
	diagramClient := diagram.NewDiagramServiceClient(diagramConn)

	log.Printf("Connecting to Analyzer Service at %s", config.AnalyzerServiceAddr)
	analyzerConn, err := grpc.NewClient(
		config.AnalyzerServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Analyzer Service: %v", err)
	}
	defer analyzerConn.Close()
	analyzerClient := analyzer.NewAnalyzerServiceClient(analyzerConn)

	// TCP 리스너 생성
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to create TCP listener: %v", err)
	}

	// gRPC 서버 생성
	grpcServer := grpc.NewServer()

	// 모든 서비스 클라이언트와 함께 핸들러 생성
	implementationHandler := handler.NewImplementationHandler(
		*config,
		planClient,
		diagramClient,
		analyzerClient,
	)
	implementation.RegisterImplementationServiceServer(grpcServer, implementationHandler)

	reflection.Register(grpcServer)

	log.Printf("Implementation Service starting on port %s", config.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
