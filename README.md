### 1. Generating code from a proto file

We will need to run the following command to generate code derived from a `.proto` file.

```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative my_proto.proto
```

