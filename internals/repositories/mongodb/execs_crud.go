package mongodb

import (
	"context"
	"time"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddExecsDBHandler(ctx context.Context, execsFromReq []*pb.Exec) ([]*pb.Exec, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	newExecs := make([]*models.Exec, len(execsFromReq))
	for i, pbExec := range execsFromReq {
		newExecs[i], err = mapPbExecToModelExec(pbExec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping exec data")
		}
		if newExecs[i] == nil {
			return nil, utils.ErrorHandler(nil, "Mapped Exec is nil")
		}
		hashedPassword, err := utils.HashPassword(newExecs[i].Password)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error hashing password")
		}
		newExecs[i].Password = hashedPassword
		currentTime := time.Now().Format(time.RFC3339)
		newExecs[i].UserCreatedAt = currentTime
		newExecs[i].InactiveStatus = false
	}

	// fmt.Println(newExecs)

	var addedExecs []*pb.Exec
	for _, exec := range newExecs {
		result, err := client.Database("school").Collection("execs").InsertOne(ctx, exec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding exec to database")
		}

		objectId, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			exec.Id = objectId.Hex()
		}

		pbExec, err := mapModelExecToPbExec(*exec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping exec data")
		}
		addedExecs = append(addedExecs, pbExec)
	}
	return addedExecs, nil
}
