package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/models"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/internals/repositories/mongodb"
	"github.com/aayushxrj/go-gRPC-api-school-mgmt/pkg/utils"
	pb "github.com/aayushxrj/go-gRPC-api-school-mgmt/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) AddExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	for _, exec := range req.GetExecs() {
		if exec.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "New exec entries should not have an ID")
		}
	}

	addedExecs, err := mongodb.AddExecsDBHandler(ctx, req.GetExecs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: addedExecs}, nil
}

func (s *Server) GetExecs(ctx context.Context, req *pb.GetExecsRequest) (*pb.Execs, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Filtering, getting the filters from the request
	filter, err := BuildFilterForTeacher(req.Exec, models.Exec{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Sorting, getting the sort options from the request
	sortOptions := BuildSortOptions(req.GetSortBy())

	// Access the database to fetch data
	execs, err := mongodb.GetExecsDBHandler(ctx, sortOptions, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: execs}, nil
}
func (s *Server) UpdateExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {
	// if err := req.Validate(); err != nil {
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	updatedExecs, err := mongodb.UpdateExecsDBHandler(ctx, req.GetExecs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: updatedExecs}, nil
}

func (s *Server) DeleteExecs(ctx context.Context, req *pb.ExecIds) (*pb.DeleteExecsConfirmation, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ids := req.GetIds()
	var execIdsToDelete []string

	for _, v := range ids {
		execIdsToDelete = append(execIdsToDelete, v)
	}

	deletedIds, err := mongodb.DeleteExecsDBHandler(ctx, execIdsToDelete)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteExecsConfirmation{Status: "Execs deleted successfully", DeletedIds: deletedIds}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.ExecLoginRequest) (*pb.ExecLoginResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	exec, err := mongodb.LoginExecDBHandler(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if exec.InactiveStatus {
		return nil, status.Error(codes.Unauthenticated, "Account is inactive")
	}

	err = utils.VerifyPassword(req.GetPassword(), exec.Password)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Incorrect password")
	}

	tokenString, err := utils.SignToken(exec.Id, exec.Username, exec.Role)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error generating auth token")
	}

	return &pb.ExecLoginResponse{
		Status: true,
		Token:  tokenString,
	}, nil
}

func (s *Server) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := mongodb.UpdatePasswordExecDBHandler(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if token == "" {
		return nil, status.Error(codes.Internal, "Failed to update password")
	}

	return &pb.UpdatePasswordResponse{
		PasswordUpdated: true,
		Token:           token,
	}, nil
}

func (s *Server) DeactivateUser(ctx context.Context, req *pb.ExecIds) (*pb.Confirmation, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := mongodb.DeactivateUserDBHandler(ctx, req.GetIds())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.ModifiedCount == 0 {
		return &pb.Confirmation{
			Confirmation: false,
		}, nil
	}

	return &pb.Confirmation{
		Confirmation: true,
	}, nil
}

func (s *Server) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	email := req.GetEmail()
	message, err := mongodb.ForgotPasswordExecDBHandler(ctx, email)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ForgotPasswordResponse{
		Confirmation: true,
		Message:      message,
	}, nil
}

func (s *Server) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.Confirmation, error) {
	token := req.GetResetCode()

	if req.GetNewPassword() != req.GetConfirmPassword() {
		return nil, status.Error(codes.InvalidArgument, "Passwords do not match")
	}

	bytes, err := hex.DecodeString(token)
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}

	hashedToken := sha256.Sum256(bytes)
	tokenInDb := hex.EncodeToString(hashedToken[:])

	err = mongodb.ResetPasswordDBHandler(ctx, tokenInDb, req.GetNewPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Confirmation{
		Confirmation: true,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *pb.EmptyRequest) (*pb.ExecLogoutResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata found")
	}

	val, ok := md["authorization"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized Access")
	}

	token := strings.TrimPrefix(val[0], "Bearer ")

	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized Access")
	}

	expiryTimeStamp := ctx.Value(utils.ContextKey("expiresAt"))
	expiryTimeStr := fmt.Sprintf("%v", expiryTimeStamp)

	expiryTimeInt, err := strconv.ParseInt(expiryTimeStr, 10, 64)
	if err != nil {
		utils.ErrorHandler(err, "")
		return nil, status.Error(codes.Internal, "internal error")
	}

	expirytime := time.Unix(expiryTimeInt, 0)

	utils.JwtStore.AddToken(token, expirytime)

	return &pb.ExecLogoutResponse{
		LoggedOut: true,
	}, nil
}
