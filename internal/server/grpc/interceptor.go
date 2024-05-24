package grpc

import (
	"context"
	"google.golang.org/grpc"
	"log"
)

func RequestLoggingInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	log.Printf("DEBUG. incoming gRPC request: %v", req)
	return handler(ctx, req)
}
