package main

import (
	"fmt"
	"log"
	"net"
	"net/url"

	_ "ariga.io/atlas-provider-gorm/gormschema"

	"codev42-plan/configs"
	"codev42-plan/handler"
	"codev42-plan/pb"
	"codev42-plan/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Load configuration
	config, err := configs.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect to MySQL
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MySQLUser,
		url.QueryEscape(config.MySQLPassword),
		config.MySQLHost,
		config.MySQLPort,
		config.MySQLDB,
	)

	log.Printf("Connecting to MySQL: %s@%s:%s/%s",
		config.MySQLUser,
		config.MySQLHost,
		config.MySQLPort,
		config.MySQLDB,
	)

	rdbConnection, err := storage.NewRDBConnection(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer rdbConnection.Close()

	log.Println("Successfully connected to MySQL")

	// 3. Create TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to create TCP listener: %v", err)
	}

	// 4. Create gRPC server
	grpcServer := grpc.NewServer()

	// 5. Register Plan Service
	planHandler := handler.NewPlanHandler(*config, rdbConnection)
	pb.RegisterPlanServiceServer(grpcServer, planHandler)

	// 6. Register reflection (for debugging with grpcurl/evans)
	reflection.Register(grpcServer)

	// 7. Start server
	log.Printf("Plan Service starting on port %s", config.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
