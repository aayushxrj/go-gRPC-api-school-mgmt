package handlers

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/repositories/mongodb"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
	filter, err := buildFilterForTeacher(req.Teacher)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Sorting, getting the sort options from the request
	sortOptions := buildSortOptions(req.GetSortBy())

	// Access the databse to fetch data
	teachers, err := mongodb.GetTeachersDBHandler(ctx, sortOptions, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: teachers}, nil
}

func buildFilterForTeacher(teacherObj *pb.Teacher) (bson.M, error) {
	filter := bson.M{}

	if teacherObj == nil {
		return filter, nil
	}

	var modelTeacher models.Teacher
	modelVal := reflect.ValueOf(&modelTeacher).Elem()
	modelType := modelVal.Type()

	reqVal := reflect.ValueOf(teacherObj).Elem()
	reqType := reqVal.Type()

	for i := 0; i < reqVal.NumField(); i++ {
		fieldVal := reqVal.Field(i)
		fieldName := reqType.Field(i).Name

		if fieldVal.IsValid() && !fieldVal.IsZero() {
			modelField := modelVal.FieldByName(fieldName)
			if modelField.IsValid() && modelField.CanSet() {
				modelField.Set(fieldVal)
			}
		}
	}

	// iterate over the modelTeacher to build the filter
	for i := 0; i < modelVal.NumField(); i++ {
		fieldVal := modelVal.Field(i)
		// fieldName := modelType.Field(i).Name

		if fieldVal.IsValid() && !fieldVal.IsZero() {
			bsonTag := modelType.Field(i).Tag.Get("bson")
			bsonTag = strings.TrimSuffix(bsonTag, ",omitempty")
			if bsonTag == "_id" {
				objID, err := primitive.ObjectIDFromHex(teacherObj.Id)
				if err != nil {
					return nil, utils.ErrorHandler(err, "Invalid ID format")
				}
				filter[bsonTag] = objID
			} else {
				filter[bsonTag] = fieldVal.Interface().(string)
			}
			// filter[bsonTag] = fieldVal.Interface().(string)
		}
	}

	fmt.Println("Filter:", filter)
	return filter, nil
}

func buildSortOptions(sortFields []*pb.SortField) bson.D {
	var sortOptions bson.D

	for _, sortField := range sortFields {
		order := 1
		if sortField.GetOrder() == pb.Order_DESC {
			order = -1
		}
		sortOptions = append(sortOptions, bson.E{Key: sortField.Field, Value: order})
	}
	fmt.Println("Sort Options:", sortOptions)
	return sortOptions
}
