package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/api/handlers"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)


func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterTeachersServiceServer(s, &handlers.Server{})
	pb.RegisterStudentsServiceServer(s, &handlers.Server{})
	pb.RegisterExecsServiceServer(s, &handlers.Server{})

	reflection.Register(s)

	// go get github.com/joho/godotenv
	port := os.Getenv("SERVER_PORT")

	fmt.Printf("Server is running on port %s\n", port)
	
	lis,err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}


}