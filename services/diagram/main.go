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
	config, err := configs.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Diagram Service configuration loaded")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to create TCP listener: %v", err)
	}

	grpcServer := grpc.NewServer()

	diagramHandler := handler.NewDiagramHandler(*config)
	pb.RegisterDiagramServiceServer(grpcServer, diagramHandler)

	reflection.Register(grpcServer)

	log.Printf("Diagram Service starting on port %s", config.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
