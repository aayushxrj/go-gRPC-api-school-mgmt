package mongodb

import (
	"context"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
