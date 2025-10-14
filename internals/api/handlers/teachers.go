package handlers

import (
	"context"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/repositories/mongodb"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	for _, teacher := range req.GetTeachers() {
		if teacher.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "New teacher entries should not have an ID")
		}
	}

	addedTeachers, err := mongodb.AddTeachersDBHandler(ctx, req.GetTeachers())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: addedTeachers}, nil
}

func (s *Server) GetTeachers(ctx context.Context, req *pb.GetTeachersRequest) (*pb.Teachers, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Filtering, getting the filters from the request
	filter, err := BuildFilterForTeacher(req.Teacher, models.Teacher{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Sorting, getting the sort options from the request
	sortOptions := BuildSortOptions(req.GetSortBy())

	// Access the databse to fetch data
	teachers, err := mongodb.GetTeachersDBHandler(ctx, sortOptions, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: teachers}, nil
}

func (s *Server) UpdateTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	updatedTeachers, err := mongodb.UpdateTeachersDBHandler(ctx, req.GetTeachers())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: updatedTeachers}, nil
}

// rpc DeleteTeachers (TeacherIds) returns (DeleteTeachersConfirmation);

// message DeleteTeachersConfirmation {
//     string status = 1;
//     repeated string deleted_ids = 2;
// }

// message TeacherId {
//     string id = 1 [(validate.rules).string = {min_len: 24, max_len: 24, pattern: "^[a-fA-F0-9]{24}$"}];
// }

// message TeacherIds {
//     repeated TeacherId ids = 1 [(validate.rules).repeated = {min_items: 1}];
// }

func (s *Server) DeleteTeachers(ctx context.Context, req *pb.TeacherIds) (*pb.DeleteTeachersConfirmation, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ids := req.GetIds()
	var teacherIdsToDelete []string

	for _, v := range ids {
		teacherIdsToDelete = append(teacherIdsToDelete, v.GetId())
	}

	deletedIds, err := mongodb.DeleteTeachersDBHandler(ctx, teacherIdsToDelete)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteTeachersConfirmation{Status: "Teachers deleted successfully", DeletedIds: deletedIds}, nil
}