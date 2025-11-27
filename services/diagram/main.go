package main

import (
	"fmt"
	"log"
	"net"

	"codev42-diagram/configs"
	"codev42-diagram/handler"
	"codev42-diagram/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Load configuration
	config, err := configs.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Diagram Service configuration loaded")

	// 2. Create TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to create TCP listener: %v", err)
	}

	// 3. Create gRPC server
	grpcServer := grpc.NewServer()

	// 4. Register Diagram Service
	diagramHandler := handler.NewDiagramHandler(*config)
	pb.RegisterDiagramServiceServer(grpcServer, diagramHandler)

	// 5. Register reflection (for debugging)
	reflection.Register(grpcServer)

	// 6. Start server
	log.Printf("Diagram Service starting on port %s", config.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
