package grpcserver

import (
	"log"
	"net"

	"github.com/sehwan505/codev42/pb"
	"google.golang.org/grpc"
)

func RunGRPCServer(port string) error {
	grpcServer := grpc.NewServer()
	agentService := NewAgentServiceServer()

	pb.RegisterAgentServiceServer(grpcServer, agentService)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
		return err
	}

	log.Println("Starting gRPC server on port", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
		return err
	}
	return nil
}
