package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"codev42/services/agent/configs"
	"codev42/services/agent/handler"
	pb "codev42/services/agent/pb"
	"codev42/services/agent/storage"
	"codev42/services/agent/storage/repo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// func connectMongo() *mongo.Client {
// 	mongo := db.Connect("POST_MONGO_URL")
// 	err := mongo.Ping(context.Background(), readpref.Primary())
// 	if err != nil {
// 		log.Println("Couldn't connect to the Mongo", err)
// 	} else {
// 		log.Println("Mongo Connected!")
// 	}

// 	return mongo
// }

type VectorDB interface {
	InitCollection(ctx context.Context, collectionName string, vectorDim int) error
	InsertEmbedding(ctx context.Context, collectionName string, id string, embedding []float32) error
	SearchByVector(ctx context.Context, collectionName string, searchVector []float32, topK int) ([]int64, error)
	Close() error
}

func main() {
	// mongo := connectMongo()
	// defer mongo.Disconnect(context.Background())
	config, err := configs.GetConfig()
	if err != nil {
		log.Fatalf("Couldn't get config %v", err)
	}
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
		config.MySQLPassword,
		config.MySQLHost,
		config.MySQLPort,
		config.MySQLDB,
	)
	_, err = storage.NewDBConnection(dsn)
	if err != nil {
		log.Fatalf("Couldn't connect to MySQL %v", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
	if err != nil {
		log.Fatalf("Couldn't create connection tcp %v", err)
	}

	agentHandler := &handler.AgentHandler{Config: *config}
	grpcServer := grpc.NewServer()
	pb.RegisterAgentServiceServer(grpcServer, agentHandler)
	reflection.Register(grpcServer)
	log.Printf("Server start at port %s", config.GRPCPort)

	grpcServer.Serve(listener)
}
