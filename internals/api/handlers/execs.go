package handlers

import (
	"context"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/repositories/mongodb"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	for _, exec := range req.GetExecs() {
		if exec.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "New exec entries should not have an ID")
		}
	}

	addedExecs, err := mongodb.AddExecsDBHandler(ctx, req.GetExecs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: addedExecs}, nil
}

func (s *Server) GetExecs(ctx context.Context, req *pb.GetExecsRequest) (*pb.Execs, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Filtering, getting the filters from the request
	filter, err := BuildFilterForTeacher(req.Exec, models.Exec{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Sorting, getting the sort options from the request
	sortOptions := BuildSortOptions(req.GetSortBy())

	// Access the database to fetch data
	execs, err := mongodb.GetExecsDBHandler(ctx, sortOptions, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: execs}, nil
}
func (s *Server) UpdateExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {
	// if err := req.Validate(); err != nil {
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	updatedExecs, err := mongodb.UpdateExecsDBHandler(ctx, req.GetExecs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: updatedExecs}, nil
}
