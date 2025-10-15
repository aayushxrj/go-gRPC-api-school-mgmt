package handlers

import (
	"context"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
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

func (s *Server) GetStudents(ctx context.Context, req *pb.GetStudentsRequest) (*pb.Students, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Filtering, getting the filters from the request
	filter, err := BuildFilterForTeacher(req.Student, models.Student{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Sorting, getting the sort options from the request
	sortOptions := BuildSortOptions(req.GetSortBy())

	// for pagination
	pageNumber := req.GetPageNumber()
	pageSize := req.GetPageSize()

	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Access the database to fetch data
	students, err := mongodb.GetStudentsDBHandler(ctx, sortOptions, filter, pageNumber, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: students}, nil
}
func (s *Server) UpdateStudents(ctx context.Context, req *pb.Students) (*pb.Students, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	updatedStudents, err := mongodb.UpdateStudentsDBHandler(ctx, req.GetStudents())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: updatedStudents}, nil
}
func (s *Server) DeleteStudents(ctx context.Context, req *pb.StudentIds) (*pb.DeleteStudentsConfirmation, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ids := req.GetIds()
	var studentIdsToDelete []string

	for _, v := range ids {
		studentIdsToDelete = append(studentIdsToDelete, v)
	}

	deletedIds, err := mongodb.DeleteStudentsDBHandler(ctx, studentIdsToDelete)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteStudentsConfirmation{Status: "Students deleted successfully", DeletedIds: deletedIds}, nil
}
