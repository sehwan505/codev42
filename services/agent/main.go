package main

import (
	"codev42/services/agent/configs"
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"time"

	_ "ariga.io/atlas-provider-gorm/gormschema"

	"codev42/services/agent/handler"
	pb "codev42/services/agent/pb"
	"codev42/services/agent/storage"
	"codev42/services/agent/storage/repo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type VectorDB interface {
	InitCollection(ctx context.Context, collectionName string, vectorDim int) error
	InsertEmbedding(ctx context.Context, collectionName string, id string, embedding []float32) error
	SearchByVector(ctx context.Context, collectionName string, searchVector []float32, topK int) ([]int64, error)
	Close() error
}

func setStorage(config *configs.Config) (VectorDB, *storage.RDBConnection) {
	useMilvus := false

	var vectorDB VectorDB
	if useMilvus {
		ctx := context.Background()
		milvusConn, err := storage.NewMilvusConnection(ctx, fmt.Sprintf("%s:%s", config.MilvusHost, config.MilvusPort))
		if err != nil {
			log.Fatalf("Couldn't connect to Milvus: %v", err)
		}
		vectorDB = repo.NewMilvusRepo(milvusConn)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pineconeConn, err := storage.NewPineconeConnection(ctx, config.PineconeApiKey)
		if err != nil {
			log.Fatalf("Couldn't connect to Pinecone: %v", err)
		}
		vectorDB = repo.NewPineconeRepo(pineconeConn)
		if err != nil {
			log.Fatalf("Couldn't create PineconeRepo: %v", err)
		}
	}

	ctx := context.Background()
	if err := vectorDB.InitCollection(ctx, "my-collection1", 384); err != nil {
		log.Fatalf("Failed to init collection: %v", err)
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MySQLUser,
		url.QueryEscape(config.MySQLPassword),
		config.MySQLHost,
		config.MySQLPort,
		config.MySQLDB,
	)
	fmt.Printf("dsn: %s\n", dsn)
	// MariaDB 연결
	rdbConnection, err := storage.NewRDBConnection(dsn)
	if err != nil {
		log.Fatalf("Couldn't connect to MySQL %v", err)
	}
	// // 자동 마이그레이션
	// rdbConnection.AutoMigrate()

	return vectorDB, rdbConnection
}

func main() {
	config, err := configs.GetConfig()
	vectorDB, rdbConnection := setStorage(config)
	if err != nil {
		log.Fatalf("Couldn't get config %v", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		log.Fatalf("Couldn't create connection tcp %v", err)
	}

	agentHandler := &handler.AgentHandler{Config: *config}
	codeHandler := &handler.CodeHandler{Config: *config, VectorDB: vectorDB, RdbConnection: rdbConnection}
	grpcServer := grpc.NewServer()
	pb.RegisterCodeServiceServer(grpcServer, codeHandler)
	pb.RegisterAgentServiceServer(grpcServer, agentHandler)
	reflection.Register(grpcServer)
	log.Printf("Server start at port %s", config.GRPCPort)
	grpcServer.Serve(listener)
	rdbConnection.Close()
}
