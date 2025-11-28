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
	config, err := configs.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Implementation Service configuration loaded")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to create TCP listener: %v", err)
	}

	grpcServer := grpc.NewServer()

	implementationHandler := handler.NewImplementationHandler(*config)
	pb.RegisterImplementationServiceServer(grpcServer, implementationHandler)

	reflection.Register(grpcServer)

	log.Printf("Implementation Service starting on port %s", config.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
