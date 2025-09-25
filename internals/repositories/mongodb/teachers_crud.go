package mongodb

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddTeachersDBHandler(ctx context.Context, teachersFromReq []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	newTeachers := make([]*models.Teacher, len(teachersFromReq))
	for i, pbTeacher := range teachersFromReq {
		newTeachers[i], err = mapPbTeacherToModelTeacher(pbTeacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping teacher data")
		}
	}

	// fmt.Println(newTeachers)

	var addedTeachers []*pb.Teacher
	for _, teacher := range newTeachers {
		result, err := client.Database("school").Collection("teachers").InsertOne(ctx, teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding teacher to database")
		}

		objectId, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			teacher.Id = objectId.Hex()
		}

		pbTeacher, err := mapModelTeacherToPbTeacher(teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping teacher data")
		}
		addedTeachers = append(addedTeachers, pbTeacher)
	}
	return addedTeachers, nil
}

func mapModelTeacherToPbTeacher(teacher *models.Teacher) (*pb.Teacher, error) {
	pbTeacher := &pb.Teacher{}
	modelVal := reflect.ValueOf(*teacher)
	pbVal := reflect.ValueOf(pbTeacher).Elem()

	for i := 0; i < modelVal.NumField(); i++ {
		modelField := modelVal.Field(i)
		modelFieldTypeName := modelVal.Type().Field(i).Name

		pbField := pbVal.FieldByName(modelFieldTypeName)
		if pbField.IsValid() && pbField.CanSet() {
			pbField.Set(modelField)
		} else {
			return nil, fmt.Errorf("field %s not found or cannot be set in pb.Teacher", modelFieldTypeName)
		}
	}
	return pbTeacher, nil
}

func mapPbTeacherToModelTeacher(pbTeacher *pb.Teacher) (*models.Teacher, error) {
	modelTeacher := models.Teacher{}
	pbVal := reflect.ValueOf(pbTeacher).Elem()
	modelVal := reflect.ValueOf(&modelTeacher).Elem()

	for i := 0; i < pbVal.NumField(); i++ {
		pbField := pbVal.Field(i)
		fieldName := pbVal.Type().Field(i).Name

		modelField := modelVal.FieldByName(fieldName)
		if modelField.IsValid() && modelField.CanSet() {
			modelField.Set(pbField)
		} else {
			return nil, fmt.Errorf("field %s not found or cannot be set in models.Teacher", fieldName)
		}
	}
	return &modelTeacher, nil
}
