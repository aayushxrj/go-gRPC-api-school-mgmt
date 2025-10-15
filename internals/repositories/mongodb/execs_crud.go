package mongodb

import (
	"context"
	"time"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func GetExecsDBHandler(ctx context.Context, sortOptions primitive.D, filter primitive.M) ([]*pb.Exec, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("execs")
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

	execs, err := DecodeEntities(ctx,
		cursor,
		func() *pb.Exec { return &pb.Exec{} },
		func() *models.Exec { return &models.Exec{} })
	if err != nil {
		return nil, err
	}
	return execs, nil
}

func UpdateExecsDBHandler(ctx context.Context, pbExecs []*pb.Exec) ([]*pb.Exec, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	var updatedExecs []*pb.Exec

	for _, exec := range pbExecs {
		if exec.Id == "" {
			return nil, utils.ErrorHandler(nil, "Exec ID is required for update")
		}

		modelExec, err := mapPbExecToModelExec(exec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping exec data")
		}

		objID, err := primitive.ObjectIDFromHex(exec.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid ID format")
		}

		// Hash password if provided
		if exec.Password != "" {
			hashed, err := utils.HashPassword(exec.Password)
			if err != nil {
				return nil, utils.ErrorHandler(err, "Error hashing password")
			}
			modelExec.Password = hashed
		}

		// Convert modelExec to BSON document
		modelDoc, err := bson.Marshal(modelExec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error preparing exec data for update")
		}

		var updateDoc bson.M
		if err := bson.Unmarshal(modelDoc, &updateDoc); err != nil {
			return nil, utils.ErrorHandler(err, "Error preparing exec data for update")
		}

		// Do not update _id
		delete(updateDoc, "_id")

		// Do not update password if it wasn't provided
		if exec.Password == "" {
			delete(updateDoc, "password")
		}

		_, err = client.Database("school").Collection("execs").UpdateOne(
			ctx,
			bson.M{"_id": objID},
			bson.M{"$set": updateDoc},
		)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error updating exec data")
		}

		updatedExec, err := mapModelExecToPbExec(*modelExec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error mapping exec data")
		}
		updatedExec.Id = exec.Id

		updatedExecs = append(updatedExecs, updatedExec)
	}
	return updatedExecs, nil
}
