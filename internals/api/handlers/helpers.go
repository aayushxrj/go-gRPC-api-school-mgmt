package handlers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
)

// func buildFilterForTeacher(object *pb.Teacher) (bson.M, error) {
func BuildFilterForTeacher(object interface{}, model interface{}) (bson.M, error) {
	filter := bson.M{}

	if object == nil {
		return filter, nil
	}

	val := reflect.ValueOf(object)
	if !val.IsValid() || (val.Kind() == reflect.Ptr && val.IsNil()) {
		return filter, nil
	}
	modelVal := val.Elem()
	modelType := modelVal.Type()

	reqVal := modelVal
	reqType := modelType

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
		fieldName := modelType.Field(i).Name

		if fieldVal.IsValid() && !fieldVal.IsZero() {
			bsonTag := modelType.Field(i).Tag.Get("bson")
			bsonTag = strings.TrimSuffix(bsonTag, ",omitempty")
			if bsonTag == "_id" {
				// objID, err := primitive.ObjectIDFromHex(object.Id)
				objID, err := primitive.ObjectIDFromHex(reqVal.FieldByName(fieldName).Interface().(string))
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

func BuildSortOptions(sortFields []*pb.SortField) bson.D {
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
