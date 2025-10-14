package handlers

import (
	"context"

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
