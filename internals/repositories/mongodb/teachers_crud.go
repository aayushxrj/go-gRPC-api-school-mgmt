package mongodb

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func GetTeachersDBHandler(ctx context.Context, sortOptions primitive.D, filter primitive.M) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("teachers")
	var cursor *mongo.Cursor
	if len(sortOptions) < 1 {
		cursor, err = coll.Find(ctx, filter)
	} else {
		cursor, err = coll.Find(ctx, filter, options.Find().SetSort(sortOptions))
	}
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}
	defer cursor.Close(ctx)

	teachers, err := decodeEntities(ctx,
		cursor,
		func() *pb.Teacher { return &pb.Teacher{} },
		func() *models.Teacher { return &models.Teacher{} })
	if err != nil {
		return nil, err
	}
	return teachers, nil
}

func decodeEntities[T any, M any](ctx context.Context, cursor *mongo.Cursor, newEntity func() *T, newModel func() *M) ([]*T, error) {
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
