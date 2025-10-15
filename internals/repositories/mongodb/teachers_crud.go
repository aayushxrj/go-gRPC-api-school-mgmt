package mongodb

import (
	"context"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

		pbTeacher, err := mapModelTeacherToPbTeacher(*teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping teacher data")
		}
		addedTeachers = append(addedTeachers, pbTeacher)
	}
	return addedTeachers, nil
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

	teachers, err := DecodeEntities(ctx,
		cursor,
		func() *pb.Teacher { return &pb.Teacher{} },
		func() *models.Teacher { return &models.Teacher{} })
	if err != nil {
		return nil, err
	}
	return teachers, nil
}

func UpdateTeachersDBHandler(ctx context.Context, pbTeachers []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Database connection error")
	}
	defer client.Disconnect(ctx)

	var updatedTeachers []*pb.Teacher

	for _, teacher := range pbTeachers {

		if teacher.Id == "" {
			return nil, status.Error(codes.InvalidArgument, "Teacher ID is required for update")
		}

		modelTeacher, err := mapPbTeacherToModelTeacher(teacher)
		if err != nil {
			return nil, status.Error(codes.Internal, "Error mapping teacher data")
		}

		objID, err := primitive.ObjectIDFromHex(teacher.Id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "Invalid ID format")
		}

		// converting modelTeacher to bson Document
		modelDoc, err := bson.Marshal(modelTeacher)
		if err != nil {
			return nil, status.Error(codes.Internal, "Error preparing teacher data for update")
		}

		var updateDoc bson.M
		err = bson.Unmarshal(modelDoc, &updateDoc)
		if err != nil {
			return nil, status.Error(codes.Internal, "Error preparing teacher data for update")
		}

		// remove the _id field from the update document
		delete(updateDoc, "_id")

		_, err = client.Database("school").Collection("teachers").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, status.Error(codes.Internal, "Error updating teacher data")
		}

		updatedTeacher, err := mapModelTeacherToPbTeacher(*modelTeacher)
		if err != nil {
			return nil, status.Error(codes.Internal, "Error mapping teacher data")
		}

		updatedTeachers = append(updatedTeachers, updatedTeacher)
	}
	return updatedTeachers, nil
}

func DeleteTeachersDBHandler(ctx context.Context, teacherIdsToDelete []string) ([]string, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	objectIds := make([]primitive.ObjectID, len(teacherIdsToDelete))
	for i, id := range teacherIdsToDelete {
		if id == "" {
			return nil, utils.ErrorHandler(err, "Teacher ID is required for deletion")
		}
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid teacher ID format")
		}
		objectIds[i] = objID
	}

	filter := bson.M{"_id": bson.M{"$in": objectIds}}
	result, err := client.Database("school").Collection("teachers").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error deleting teachers from database")
	}

	if result.DeletedCount == 0 {
		return nil, status.Error(codes.NotFound, "No teachers found to delete")
	}

	deletedIds := make([]string, result.DeletedCount)

	for i, objID := range objectIds {
		deletedIds[i] = objID.Hex()
	}

	return deletedIds, nil
}

func GetStudentsByClassTeacherDBHandler(ctx context.Context, teacherId string) ([]*pb.Student, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	objID, err := primitive.ObjectIDFromHex(teacherId)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Invalid teacher ID format")
	}

	var teacher models.Teacher
	err = client.Database("school").Collection("teachers").FindOne(ctx, bson.M{"_id": objID}).Decode(&teacher)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, utils.ErrorHandler(err, "Teacher not found")
		}
		return nil, utils.ErrorHandler(err, "Error fetching teacher data")
	}

	cursor, err := client.Database("school").Collection("students").Find(ctx, bson.M{"class": teacher.Class})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error fetching students by class")
	}
	defer cursor.Close(ctx)

	students, err := DecodeEntities(ctx,
		cursor,
		func() *pb.Student { return &pb.Student{} },
		func() *models.Student { return &models.Student{} })	
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error decoding student data")
	}

	return students, nil
}
