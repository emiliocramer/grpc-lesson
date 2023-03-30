# gRPC & go, from 0 to 1
The following is a step by step guide on writing all the code within this repository.

Inside are some example services that can be exposed by a gRPC server running locally. Services include
1. Request -> Response
2. Stream -> Response
3. Request -> Stream
4. Stream -> Stream

#### And more...

there is a client that interacts with all these services located in `client.go` and a server in `server.go`.

Both of these get ran at the same time in `main.go` at once, _this is probably not the best way of doing it, but for sake of ease... thats how it is._

## Steps

### 1. Creating a proto file

We start by creating a proto file that contains `messages` and `services`.  Ours looks
like the following. We must specify a go output module if we want to work in go moving forward

```
syntax = "proto3";

package proto;

option go_package = "./main";

service CalculatorService {
  rpc TimesTen(Message) returns (Response) {}
}

message Message {
  string say = 1;
}

message Response {
  string say = 1;
}
```

### 2. Generating code from a proto file

We will need to run the following command to generate code derived from a `.proto` file.

```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative my_proto.proto
```

### 3. Implementing a gRPC server in Go

### 3.1 Defining our server

In the `serve.go` file we initialize the actual server on its respective port.

```
func ServeFunc() {
    
    // Create the listener
	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	// Create the server 
	grpcServer := grpc.NewServer()
	proto.RegisterCalculatorServiceServer(grpcServer, &server{})
	
	// Launch the server
	fmt.Println("Calculator sercer is running on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
```

### 3.2 Defining the services
We now need to make a server in go that implements the services generated in step `#2`.

In our case this gets done in the `serve.go` file. For each service exposed we would want to write something along the lines of

```
func (s *server) TimesTen(ctx context.Context, req *proto.Message) (*proto.Response, error) {

	// Get the request
	numberString := req.GetSay()
	numberAsInt, err := strconv.Atoi(numberString)
	if err != nil {
		return nil, fmt.Errorf("unable to convert request into an integer, try using a real number")
	}

	// Construct the response
	result := string(rune(numberAsInt * 10))
	res := &proto.Response{
		Say: result,
	}

	// Return the response
	return res, nil
}
```

### 4. Create a client to commununicate with our server

In a new file `client.go` we create the client who will reach our server and target the endpoints we've exposed

We do so like the following

### 4.1 Creating the client

The client itself looks something like this

```
func ClientFunc() {
	fmt.Println("Calculator client starting...")

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed tp connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewCalculatorServiceClient(conn)
	doTimesTen(client)
}
```

### 4.2 Creating the function that targets the service

Once a client that reaches our `localhost` is made, we need a function that targets the actual individual service

That is done so like the following

```
func doTimesTen(c proto.CalculatorServiceClient) {
	fmt.Println("Starting")

	req := &proto.Message{
		Say: 100,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.TimesTen(ctx, req)
	if err != nil {
		log.Fatalf("Error calling TimesTen: %v", err)
	}
	log.Printf("Times Ten result: %v", res.GetSay())
}
```

### 5. Creating a Streaming service for our server

This is done by exposing a new service in our proto file with a slight change

```
  rpc Decompose(DecomposeRequest) returns (stream DecomposeResponse) {}
```

We then have to define a way to handle clients hitting this service with a new function to handle streams

```
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
```

### 5.1 Creating the client that hits and receives our streaming service

Similarly to how we defined the function that hits our first service, we have to make one for our streaming-output service.

The only change is putting the response received inside of a for-loop

```
func doDecompose(c proto.CalculatorServiceClient) {
	fmt.Println("\nStarting to decompose a number...")

	// Populate the request
	req := &proto.DecomposeRequest{
		Number: 4780,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
```


### 6. Creating a service with a stream input

This will follow a similar process as before, except when we are hitting our service, we are sending it a stream, 
and as a response we simply get a value

#### 6.1 Defining the service in proto

```
  rpc ComputeAverage(stream ComputeAverageRequest) returns (ComputeAverageResponse) {}
```

#### 6.2 Defining a func for our server

```
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
```

#### 6.3 Defining a client side request

````
func doComputeAverage(c proto.CalculatorServiceClient) {
	fmt.Println("\nStarting to calculate average...")

	numbers := []float32{1, 3, 5, 7}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
````

### 7. Creating an interceptor

An interceptor is used to intercept and modify gRPC requests and responses.
They can be used for logging, auth, or rate limiting.

To define an interceptor we create a new interceptor in `interceptors.go`

#### 7.0.0 Logging Interceptor
````
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	fmt.Printf("Request: %v\n", req)

	resp, err = handler(ctx, req)
	fmt.Printf("Response: %v\n", resp)

	return resp, err
}
````

#### 7.0.1 Stream Interceptor
````
func LoggingStreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	fmt.Printf("Streaming Request: %v\n", info)
	err := handler(srv, stream)

	fmt.Printf("Streaming Response: %v\n", info)
	return err
}
````

This interceptor is simple and is used simply for logging.

#### 7.1 Defining the interceptor on server creation

When creating the server in `server.go`, add the creation of the interceptor we have just created

```
// Create the server
grpcServer := grpc.NewServer(

    // Define the interceptor
    grpc.UnaryInterceptor(LoggingInterceptor),
    )
```

### 7.2 Making sure we hit the interceptor

In our `serve.go` we can overwrite with the following to create a custom server stream.

```
type customServerStream struct {
	grpc.ServerStream
}

func (c *customServerStream) RecvMsg(m interface{}) error {
	return c.ServerStream.RecvMsg(m)
}
```
