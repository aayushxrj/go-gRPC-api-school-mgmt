package handlers

import (
	"context"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/repositories/mongodb"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddStudents(ctx context.Context, req *pb.Students) (*pb.Students, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	for _, student := range req.GetStudents() {
		if student.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "New student entries should not have an ID")
		}
	}

	addedStudents, err := mongodb.AddStudentsDBHandler(ctx, req.GetStudents())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: addedStudents}, nil
}
