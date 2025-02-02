package main

import (
	"fmt"
	"log"
	"net"

	"codev42-gitcontrol/config"
	pb "codev42-gitcontrol/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Couldn't get config %v", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		log.Fatalf("Couldn't create connection tcp %v", err)
	}

	gitHandler := &handler.gitHandler{Config: *config}
	grpcServer := grpc.NewServer()
	pb.RegisterCodeServiceServer(grpcServer, gitHandler)
	reflection.Register(grpcServer)
	log.Printf("Server start at port %s", config.GRPCPort)
	grpcServer.Serve(listener)
}
