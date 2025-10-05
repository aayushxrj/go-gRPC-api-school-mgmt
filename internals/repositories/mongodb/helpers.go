package mongodb

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"go.mongodb.org/mongo-driver/mongo"
)

func MapModelTeacherToPbTeacher(teacher *models.Teacher) (*pb.Teacher, error) {
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

func MapPbTeacherToModelTeacher(pbTeacher *pb.Teacher) (*models.Teacher, error) {
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

func DecodeEntities[T any, M any](ctx context.Context, cursor *mongo.Cursor, newEntity func() *T, newModel func() *M) ([]*T, error) {
	var entities []*T
	for cursor.Next(ctx) {
		model := newModel()
		err := cursor.Decode(&model)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal Error")
		}
		entity := newEntity()
		modelVal := reflect.ValueOf(model).Elem()
		pbVal := reflect.ValueOf(entity).Elem()

		for i := 0; i < modelVal.NumField(); i++ {
			modelField := modelVal.Field(i)
			modelFieldName := modelVal.Type().Field(i).Name

			pbField := pbVal.FieldByName(modelFieldName)
			if pbField.IsValid() && pbField.CanSet() {
				pbField.Set(modelField)
			}
		}
		entities = append(entities, entity)
	}

	err := cursor.Err()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	return entities, nil
}
