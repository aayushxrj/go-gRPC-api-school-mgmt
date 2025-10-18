package interceptors

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ResponseTimeInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	log.Println("ResponseTimeInterceptor invoked")
	// Record the start time
	start := time.Now()

	// Call the handler to proceed with the normal execution of the RPC
	resp, err := handler(ctx, req)

	// Calculate the response time
	elapsed := time.Since(start)

	// Log the request details with duration
	st, _ := status.FromError(err)
	fmt.Printf("Method: %s, Status: %s, Response Time: %s\n", info.FullMethod, st.Code(), elapsed)

	md := metadata.Pairs("X-Response-Time", elapsed.String())
	grpc.SendHeader(ctx, md)

	log.Println("Sending response from ResponseTimeInterceptor")
	return resp, err
}
