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

// func MapModelTeacherToPbTeacher(teacher *models.Teacher) (*pb.Teacher, error) {
// 	pbTeacher := &pb.Teacher{}
// 	modelVal := reflect.ValueOf(*teacher)
// 	pbVal := reflect.ValueOf(pbTeacher).Elem()

// 	for i := 0; i < modelVal.NumField(); i++ {
// 		modelField := modelVal.Field(i)
// 		modelFieldTypeName := modelVal.Type().Field(i).Name

// 		pbField := pbVal.FieldByName(modelFieldTypeName)
// 		if pbField.IsValid() && pbField.CanSet() {
// 			pbField.Set(modelField)
// 		} else {
// 			return nil, fmt.Errorf("field %s not found or cannot be set in pb.Teacher", modelFieldTypeName)
// 		}
// 	}
// 	return pbTeacher, nil
// }

// func MapPbTeacherToModelTeacher(pbTeacher *pb.Teacher) (*models.Teacher, error) {
// 	modelTeacher := models.Teacher{}
// 	pbVal := reflect.ValueOf(pbTeacher).Elem()
// 	modelVal := reflect.ValueOf(&modelTeacher).Elem()

// 	for i := 0; i < pbVal.NumField(); i++ {
// 		pbField := pbVal.Field(i)
// 		fieldName := pbVal.Type().Field(i).Name

// 		modelField := modelVal.FieldByName(fieldName)
// 		if modelField.IsValid() && modelField.CanSet() {
// 			modelField.Set(pbField)
// 		} else {
// 			return nil, fmt.Errorf("field %s not found or cannot be set in models.Teacher", fieldName)
// 		}
// 	}
// 	return &modelTeacher, nil
// }


func mapModelToPb[P any, M any](model M, newPb func() *P) (*P, error) {
	pbStruct := newPb()
	modelVal := reflect.ValueOf(model)
	pbVal := reflect.ValueOf(pbStruct).Elem()

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

	return pbStruct, nil
}

func mapModelTeacherToPbTeacher(teacherModel models.Teacher) (*pb.Teacher, error) {
	return mapModelToPb(teacherModel, func() *pb.Teacher { return &pb.Teacher{} })
}

func mapModelStudentToPbStudent(studentModel models.Student) (*pb.Student, error) {
	return mapModelToPb(studentModel, func() *pb.Student { return &pb.Student{} })
}

func mapModelExecToPbExec(execModel models.Exec) (*pb.Exec, error) {
	return mapModelToPb(execModel, func() *pb.Exec { return &pb.Exec{} })
}

func mapPbToModel[P any, M any](pbStruct P, newModel func() *M) (*M, error) {

	modelStruct := newModel()
	pbVal := reflect.ValueOf(pbStruct).Elem()
	modelVal := reflect.ValueOf(modelStruct).Elem()

	for i := 0; i < pbVal.NumField(); i++ {
		pbField := pbVal.Field(i)
		fieldName := pbVal.Type().Field(i).Name

		modelField := modelVal.FieldByName(fieldName)
		if modelField.IsValid() && modelField.CanSet() {
			modelField.Set(pbField)
		} 
	}

	return modelStruct, nil
}

func mapPbTeacherToModelTeacher(pbTeacher *pb.Teacher) (*models.Teacher, error) {
	return mapPbToModel(pbTeacher, func() *models.Teacher { return &models.Teacher{} })
}

func mapPbStudentToModelStudent(pbStudent *pb.Student) (*models.Student, error) {
	return mapPbToModel(pbStudent, func() *models.Student { return &models.Student{} })
}

func mapPbExecToModelExec(pbExec *pb.Exec) (*models.Exec, error) {
	return mapPbToModel(pbExec, func() *models.Exec { return &models.Exec{} })
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
