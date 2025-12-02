package main

import (
	"fmt"
	"log"
	"net"

	"codev42-analyzer/configs"
	"codev42-analyzer/handler"
	"codev42-analyzer/proto/analyzer"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := configs.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Analyzer Service configuration loaded")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to create TCP listener: %v", err)
	}

	grpcServer := grpc.NewServer()

	analyzerHandler := handler.NewAnalyzerHandler(*config)
	analyzer.RegisterAnalyzerServiceServer(grpcServer, analyzerHandler)

	reflection.Register(grpcServer)

	log.Printf("Analyzer Service starting on port %s", config.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
