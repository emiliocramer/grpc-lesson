package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
)

func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	fmt.Println("\nHit interceptor")
	fmt.Printf("Interceptor Request: %v\n", req)
	resp, err = handler(ctx, req)
	fmt.Printf("Interceptor Response: %v\n", resp)

	return resp, err
}

func LoggingStreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	fmt.Printf("\nHitting stream interceptor: %v\n", info)
	customStream := &customServerStream{
		stream,
	}
	err := handler(srv, customStream)

	fmt.Printf("Interceptor Streaming Response: %v\n", info)
	return err
}
