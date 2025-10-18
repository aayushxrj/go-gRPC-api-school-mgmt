package mongodb

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

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
func DeleteExecsDBHandler(ctx context.Context, execIdsToDelete []string) ([]string, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	objectIds := make([]primitive.ObjectID, len(execIdsToDelete))
	for i, id := range execIdsToDelete {
		if id == "" {
			return nil, utils.ErrorHandler(nil, "Exec ID is required for deletion")
		}
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid exec ID format")
		}
		objectIds[i] = objID
	}

	filter := bson.M{"_id": bson.M{"$in": objectIds}}
	result, err := client.Database("school").Collection("execs").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error deleting execs from database")
	}

	if result.DeletedCount == 0 {
		return nil, utils.ErrorHandler(nil, "No execs found to delete")
	}

	return execIdsToDelete, nil
}

func LoginExecDBHandler(ctx context.Context, req *pb.ExecLoginRequest) (*models.Exec, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	filter := bson.M{"username": req.GetUsername()}
	var exec models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, filter).Decode(&exec)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, utils.ErrorHandler(err, "Exec not found")
		}
		return nil, utils.ErrorHandler(err, "Error fetching exec data")
	}

	return &exec, nil
}

func UpdatePasswordExecDBHandler(ctx context.Context, req *pb.UpdatePasswordRequest) (string, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return "", utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	objId, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return "", utils.ErrorHandler(err, "Invalid ID format")
	}

	filter := bson.M{"_id": objId}
	var exec models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, filter).Decode(&exec)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", utils.ErrorHandler(err, "Exec not found")
		}
		return "", utils.ErrorHandler(err, "Error fetching exec data")
	}

	if exec.InactiveStatus {
		return "", status.Error(codes.Unauthenticated, "Account is inactive")
	}

	err = utils.VerifyPassword(req.GetCurrentPassword(), exec.Password)
	if err != nil {
		return "", utils.ErrorHandler(err, "Incorrect old password")
	}

	hashedNewPassword, err := utils.HashPassword(req.GetNewPassword())
	if err != nil {
		return "", utils.ErrorHandler(err, "Error hashing new password")
	}

	update := bson.M{
		"$set": bson.M{
			"password":            hashedNewPassword,
			"password_changed_at": time.Now().Format(time.RFC3339),
		},
	}

	_, err = client.Database("school").Collection("execs").UpdateOne(ctx, filter, update)
	if err != nil {
		return "", utils.ErrorHandler(err, "Error updating password")
	}

	token, err := utils.SignToken(exec.Id, exec.Username, exec.Role)
	if err != nil {
		return "", utils.ErrorHandler(err, "Error generating auth token")
	}

	return token, nil
}

func DeactivateUserDBHandler(ctx context.Context, execIdsToDeactivate []string) (*mongo.UpdateResult, error) {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database connection error")
	}
	defer client.Disconnect(ctx)

	var objectIds []primitive.ObjectID
	for _, id := range execIdsToDeactivate {
		if id == "" {
			return nil, utils.ErrorHandler(nil, "Exec ID is required for deactivation")
		}
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid exec ID format")
		}
		objectIds = append(objectIds, objID)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIds}}
	update := bson.M{"$set": bson.M{"inactive_status": true}}
	res, err := client.Database("school").Collection("execs").UpdateMany(ctx, filter, update)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error deactivating execs")
	}

	return res, nil
}

func ForgotPasswordExecDBHandler(ctx context.Context, email string) (string, error) {
	client, err := CreateMongoClient(ctx)

	if err != nil {
		return "", utils.ErrorHandler(err, "Database connection error")
	}

	defer client.Disconnect(ctx)

	var exec models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, bson.M{"email": email}).Decode(&exec)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", utils.ErrorHandler(err, "Exec not found")
		}
		return "", utils.ErrorHandler(err, "Error fetching exec data")
	}

	tokenBytes := make([]byte, 32)

	_, err = rand.Read(tokenBytes)
	if err != nil {
		return "", utils.ErrorHandler(err, "Error generating reset token")
	}

	token := hex.EncodeToString(tokenBytes)
	hashedToken := sha256.Sum256(tokenBytes)
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		return "", utils.ErrorHandler(err, "Failed to send password reset email")
	}
	mins := time.Duration(duration)
	expiry := time.Now().Add(mins * time.Minute).Format(time.RFC3339)

	update := bson.M{
		"$set": bson.M{
			"password_reset_token":   hashedTokenString,
			"password_token_expires": expiry,
		},
	}
	_, err = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"email": email}, update)
	if err != nil {
		return "", utils.ErrorHandler(err, "internal error")
	}

	resetUrl := fmt.Sprintf("https://localhost:50051/execs/resetpassword/reset/%s", token)
	message := fmt.Sprintf("Forgot your password? Reset your passsword using the following link: \n%s\nPlease use the reset code:: %s along with your request to change password.\nIf you didn't request a password reset, please ignore this email.\nThis link is only valid for %v minutes.", resetUrl, token, mins)
	// subject := "Your password reset link"

	// m := mail.NewMessage()
	// m.SetHeader("From", "schooladmin@school.com")
	// m.SetHeader("To", email)
	// m.SetHeader("Subject", subject)
	// m.SetBody("text/plain", message)

	// d := mail.NewDialer("localhost", 1025, "", "")
	// err = d.DialAndSend(m)
	if err != nil {
		cleanup := bson.M{
			"$set": bson.M{
				"password_reset_token":   nil,
				"password_token_expires": nil,
			},
		}
		_, _ = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"email": email}, cleanup)
		return "", utils.ErrorHandler(err, "Could not send password reset email. Please try again")
	}
	return message, nil
}

func ResetPasswordDBHandler(ctx context.Context, tokenInDb string, newPassword string) error {
	client, err := CreateMongoClient(ctx)
	if err != nil {
		return utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	var exec models.Exec
	filter := bson.M{
		"password_reset_token": tokenInDb,
		"password_token_expires": bson.M{
			"$gt": time.Now().Format(time.RFC3339),
		},
	}
	err = client.Database("school").Collection("execs").FindOne(ctx, filter).Decode(&exec)
	if err != nil {
		return utils.ErrorHandler(err, "Invalid or expired token")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return utils.ErrorHandler(err, "internal error")
	}

	update := bson.M{
		"$set": bson.M{
			"password":               hashedPassword,
			"password_reset_token":   nil,
			"password_token_expires": nil,
			"password_changed_at":    time.Now().Format(time.RFC3339),
		},
	}
	_, err = client.Database("school").Collection("execs").UpdateOne(ctx, filter, update)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to update the password")
	}
	return nil
}