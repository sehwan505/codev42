package main

import (
	"fmt"
	"log"
	"net"

	"github.com/sehwan505/codev42/services/agent/configs"
	"github.com/sehwan505/codev42/services/agent/handler"
	pb "github.com/sehwan505/codev42/services/agent/pb"

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

func main() {
	// mongo := connectMongo()
	// defer mongo.Disconnect(context.Background())
	config, err := configs.GetConfig()
	if err != nil {
		log.Fatalf("Couldn't get config %v", err)
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
