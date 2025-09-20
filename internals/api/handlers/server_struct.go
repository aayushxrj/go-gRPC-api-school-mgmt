package handlers

import pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"

type Server struct {
	pb.UnimplementedTeachersServiceServer
	pb.UnimplementedStudentsServiceServer
	pb.UnimplementedExecsServiceServer
}