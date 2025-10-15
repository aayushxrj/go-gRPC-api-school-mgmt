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
)

func AddStudentsDBHandler(ctx context.Context, studentsFromReq []*pb.Student) ([]*pb.Student, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	newStudents := make([]*models.Student, len(studentsFromReq))
	for i, pbStudent := range studentsFromReq {
		newStudents[i], err = mapPbStudentToModelStudent(pbStudent)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping student data")
		}
	}

	var addedStudents []*pb.Student
	for _, student := range newStudents {
		result, err := client.Database("school").Collection("students").InsertOne(ctx, student)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding student to database")
		}

		objectId, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			student.Id = objectId.Hex()
		}

		pbStudent, err := mapModelStudentToPbStudent(*student)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping student data")
		}
		addedStudents = append(addedStudents, pbStudent)
	}
	return addedStudents, nil
}

func GetStudentsDBHandler(ctx context.Context, sortOptions primitive.D, filter primitive.M, pageNumber, pageSize uint32) ([]*pb.Student, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("students")

	// Pagination
	findOptions := options.Find()
	findOptions.SetSkip(int64((pageNumber - 1) * pageSize))
	findOptions.SetLimit(int64(pageSize))

	var cursor *mongo.Cursor
	if len(sortOptions) > 0 {
		findOptions.SetSort(sortOptions)
	}

	cursor, err = coll.Find(ctx, filter, findOptions)

	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}
	defer cursor.Close(ctx)

	students, err := DecodeEntities(ctx,
		cursor,
		func() *pb.Student { return &pb.Student{} },
		func() *models.Student { return &models.Student{} },
	)
	if err != nil {
		return nil, err
	}
	return students, nil
}

func UpdateStudentsDBHandler(ctx context.Context, pbStudents []*pb.Student) ([]*pb.Student, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	var updatedStudents []*pb.Student

	for _, student := range pbStudents {
		modelStudent, err := mapPbStudentToModelStudent(student)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping student data")
		}

		objID, err := primitive.ObjectIDFromHex(student.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid ID format")
		}

		modelDoc, err := bson.Marshal(modelStudent)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error preparing student data for update")
		}

		var updateDoc bson.M
		if err := bson.Unmarshal(modelDoc, &updateDoc); err != nil {
			return nil, utils.ErrorHandler(err, "Error preparing student data for update")
		}

		delete(updateDoc, "_id")

		_, err = client.Database("school").Collection("students").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error updating student data")
		}

		updatedStudent, err := mapModelStudentToPbStudent(*modelStudent)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping student data")
		}

		updatedStudents = append(updatedStudents, updatedStudent)
	}
	return updatedStudents, nil
}
