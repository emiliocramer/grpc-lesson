package main

import (
	"context"
	"fmt"
	"github.com/emiliocramer/grpc-lesson/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"time"
)

func doTimesTen(c proto.CalculatorServiceClient) {
	fmt.Println("\nStarting multiplication by 10...")

	req := &proto.Message{
		Say: 100,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	res, err := c.TimesTen(ctx, req)
	if err != nil {
		log.Fatalf("Error calling TimesTen: %v", err)
	}
	log.Printf("Times Ten result: %v", res.GetSay())
}

func doDecompose(c proto.CalculatorServiceClient) {
	fmt.Println("\nStarting to decompose a number...")

	// Populate the request
	req := &proto.DecomposeRequest{
		Number: 4780,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	// Hit the service
	stream, err := c.Decompose(ctx, req)
	if err != nil {
		log.Fatalf("Error calling Decompose: %v", err)
	}

	for {

		// Start reading the stream until stream stops
		res, err := stream.Recv()
		if err == io.EOF {
			// End of stream
			break
		}
		if err != nil {
			log.Fatalf("Error reading from stream: %v", err)
		}

		// If no error, print out the received factor
		fmt.Printf("Received a factor: %v\n", res.GetFactor())
	}
}

func doComputeAverage(c proto.CalculatorServiceClient) {
	fmt.Println("\nStarting to calculate average...")

	numbers := []float32{1, 3, 5, 7}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	// Hit the service
	stream, err := c.ComputeAverage(ctx)
	if err != nil {
		log.Fatalf("Error calling ComputeAverage: %v", err)
	}

	// Stream out and send the numbers we want averaged
	for _, number := range numbers {
		stream.Send(&proto.ComputeAverageRequest{
			Number: number,
		})
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error receiving response from ComputeAverage: %v", err)
	}

	fmt.Printf("Computed average: %v\n", res.GetAverage())
}

func doFindMaximum(c proto.CalculatorServiceClient) {
	fmt.Println("\nStarting to find maximum in stream of numbers...")

	numbers := []int32{1, 5, 2, 53, 532, 64}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	// Connect to our service
	stream, err := c.FindMaximum(ctx)
	if err != nil {
		log.Fatalf("Error calling FindMaximum: %v", err)
	}

	waitc := make(chan struct{})

	// Send numbers to the server
	go func() {
		for _, number := range numbers {
			fmt.Printf("Sending number: %v\n", number)

			// Package number to send
			stream.Send(&proto.FindMaximumRequest{
				Number: number,
			})
			time.Sleep(time.Second)
		}
		stream.CloseSend()
	}()

	// Receive the numbers back
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				break
			}
			if err != nil {
				log.Fatalf("Error receiving from FindMaximum stream: %v", err)
				close(waitc)
			}

			fmt.Printf("Received a maximum from server: %v\n", res.GetNumber())
		}
	}()
	time.Sleep(time.Second * 100)
}

func doCalculateWithDeadline(c proto.CalculatorServiceClient) {
	fmt.Println("\nstarting to do a CalculateWithDeadline")

	// Set a deadline of 2 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := &proto.CalculateWithDeadlineRequest{
		A: 3,
		B: 4,
	}

	res, err := c.CalculateWithDeadline(ctx, req)
	if err != nil {
		log.Fatalf("Error calling CalculateWithDeadline: %v", err)
	}

	fmt.Printf("\nReceived a sum...\n%v\n", res)
}

func ClientFunc() {
	fmt.Println("Calculator client starting...")

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed tp connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewCalculatorServiceClient(conn)

	doTimesTen(client)
	doDecompose(client)
	doComputeAverage(client)

	// Commented because has to stay open
	//doFindMaximum(client)

	doCalculateWithDeadline(client)
}
