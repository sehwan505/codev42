package main

import (
	"fmt"
	"log"
	"net"

	"codev42-implementation/configs"
	"codev42-implementation/handler"
	"codev42-implementation/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Load configuration
	config, err := configs.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Implementation Service configuration loaded")

	// 2. Create TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to create TCP listener: %v", err)
	}

	// 3. Create gRPC server
	grpcServer := grpc.NewServer()

	// 4. Register Implementation Service
	implementationHandler := handler.NewImplementationHandler(*config)
	pb.RegisterImplementationServiceServer(grpcServer, implementationHandler)

	// 5. Register reflection (for debugging)
	reflection.Register(grpcServer)

	// 6. Start server
	log.Printf("Implementation Service starting on port %s", config.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
