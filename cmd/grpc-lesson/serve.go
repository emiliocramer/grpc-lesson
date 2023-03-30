package main

import (
	"context"
	"fmt"
	"github.com/emiliocramer/grpc-lesson/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"time"
)

type server struct {
	proto.CalculatorServiceServer
}

type customServerStream struct {
	grpc.ServerStream
}

func (c *customServerStream) RecvMsg(m interface{}) error {
	return c.ServerStream.RecvMsg(m)
}

func ServeFunc() {

	// Create the listener
	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create the server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(LoggingInterceptor),
		grpc.StreamInterceptor(LoggingStreamInterceptor),
	)
	proto.RegisterCalculatorServiceServer(grpcServer, &server{})

	// Launch the server
	fmt.Println("Calculator server is running on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) TimesTen(ctx context.Context, req *proto.Message) (*proto.Response, error) {

	// Get the request
	number := req.GetSay()

	// Construct the response
	result := number * 10
	res := &proto.Response{
		Say: result,
	}

	// Return the response
	return res, nil
}

func (s *server) Decompose(req *proto.DecomposeRequest, stream proto.CalculatorService_DecomposeServer) error {

	// Get the number from the request
	number := req.GetNumber()
	factor := int32(2)

	// While number not zero, stream out all the factors where there is no remainder.
	for number > 1 {
		if number%factor == 0 {
			res := &proto.DecomposeResponse{
				Factor: int32(factor),
			}

			// Stream it out if % 0
			stream.Send(res)
			number = number / factor
		} else {
			factor++
		}
	}
	return nil
}

func (s *server) ComputeAverage(stream proto.CalculatorService_ComputeAverageServer) error {
	sum := float32(0)
	count := 0

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			// If end of stream, calculate the average and send it to client
			average := sum / float32(count)
			return stream.SendAndClose(&proto.ComputeAverageResponse{
				Average: average,
			})
		}
		if err != nil {
			log.Fatalf("Error reading from stream: %v", err)
		}

		// If not end of stream, build up average values
		sum += req.GetNumber()
		count++
	}
}

func (s *server) FindMaximum(stream proto.CalculatorService_FindMaximumServer) error {
	fmt.Println("Received FindMaximum RPC")

	max := int32(0)

	for {

		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error reading from stream: %v", err)
			return err
		}

		number := req.GetNumber()
		if number > max {
			max = number
			res := &proto.FindMaximumResponse{
				Number: max,
			}
			if err := stream.Send(res); err != nil {
				log.Fatalf("Error sending data to stream: %v", err)
				return err
			}
		}

	}
}

func (s *server) CalculateWithDeadline(ctx context.Context, req *proto.CalculateWithDeadlineRequest) (*proto.CalculateWithDeadlineResponse, error) {
	fmt.Println("Received CalculateWithDeadline RPC")

	// Simulate hard task
	time.Sleep(1 * time.Second)

	a, b := req.GetA(), req.GetB()

	result := a + b

	if ctx.Err() == context.Canceled {

		// The client has canceled the request
		fmt.Println("the client has canceled the request")
		return nil, status.Error(codes.Canceled, "The client has canceled the request")
	}

	return &proto.CalculateWithDeadlineResponse{
		Result: result,
	}, nil
}
