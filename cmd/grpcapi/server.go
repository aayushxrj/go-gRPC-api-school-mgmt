package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"os"
	// "time"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/api/handlers"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/api/interceptors"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/repositories/mongodb"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	// "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

//go:embed .env
var envFile embed.FS

func loadEnvFromEmbeddedFile() {
	content, err := envFile.ReadFile((".env"))
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}

	tempFile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error creating .env file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(content)
	if err != nil {
		log.Fatalf("Error writing to temp file: %v", err)
	}

	err = tempFile.Close()
	if err != nil {
		log.Fatalf("Error writing to temp file: %v", err)
	}

	err = godotenv.Load(tempFile.Name())
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {

	// Load env
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }
	loadEnvFromEmbeddedFile()

	// cert := os.Getenv("CERT_FILE")
	// key := os.Getenv("KEY_FILE")

	// creds, err := credentials.NewServerTLSFromFile(cert, key)
	// if err != nil {
	// 	log.Fatalf("Failed to load TLS credentials: %v", err)
	// }

	// Connect MongoDB
	client, err := mongodb.CreateMongoClient(context.Background())
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Not using while benchmarking
	// r := interceptors.NewRateLimiter(50, time.Minute)
	// s := grpc.NewServer(grpc.ChainUnaryInterceptor(r.RateLimitInterceptor, interceptors.ResponseTimeInterceptor, interceptors.AuthenticationInterceptor), grpc.Creds(creds))

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors.ResponseTimeInterceptor, interceptors.AuthenticationInterceptor))

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
