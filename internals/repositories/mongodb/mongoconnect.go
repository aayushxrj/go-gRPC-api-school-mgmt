package mongodb

import (
	"context"
	"fmt"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateMongoClient(ctx context.Context) (*mongo.Client, error) {
	// client, err := mongo.Connect(ctx, options.Client().ApplyURI("username:password@mongodb://localhost:27017"))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to MongoDB")
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error pinging MongoDB")
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}
