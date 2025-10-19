package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/api/handlers"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/api/interceptors"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/repositories/mongodb"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	// Load env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Connect MongoDB
	client, err := mongodb.CreateMongoClient(context.Background())
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	defer client.Disconnect(context.Background())

	r := interceptors.NewRateLimiter(5, time.Minute)
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(r.RateLimitInterceptor, interceptors.ResponseTimeInterceptor, interceptors.AuthenticationInterceptor))

	pb.RegisterTeachersServiceServer(s, &handlers.Server{})
	pb.RegisterStudentsServiceServer(s, &handlers.Server{})
	pb.RegisterExecsServiceServer(s, &handlers.Server{})

	reflection.Register(s)

	// go get github.com/joho/godotenv
	port := os.Getenv("SERVER_PORT")

	go utils.JwtStore.CleanUpExpiredTokens()

	fmt.Printf("Server is running on port %s\n", port)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
