# Simple gRPC Example

This is a simple gRPC example with a Hello World service.
│ go clean -cache -modcache -testcache

## Service Definition

- **Service**: Greeter
- **Method**: SayHello
- **Request**: HelloRequest with name field
- **Response**: HelloReply with message field

## Files

- `helloworld.proto` - Protocol buffer definition
- `helloworld/helloworld.pb.go` - Go gRPC service definitions
- `server.go` - gRPC server implementation
- `client.go` - gRPC client implementation

## Generate gRPC Code

To generate the Go gRPC code from the proto file using protoc, run these commands:

### Prerequisites

Install the required tools:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Generate Code

**在 grpc-simple 目录下执行:**

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       helloworld.proto
```

**绝对路径版本:**

```bash
protoc --go_out=/Users/zack/Desktop/blockchain-go/grpc-simple \
       --go_opt=paths=source_relative \
       --go-grpc_out=/Users/zack/Desktop/blockchain-go/grpc-simple \
       --go-grpc_opt=paths=source_relative \
       /Users/zack/Desktop/blockchain-go/grpc-simple/helloworld.proto
```

**生成的文件:**

- `helloworld/helloworld.pb.go` (消息类型)
- `helloworld/helloworld_grpc.pb.go` (gRPC 服务定义)

### Alternative: Single file generation

```bash
# Generate both in one command
protoc --go_out=. --go-grpc_out=. \
       --go_opt=paths=source_relative \
       --go-grpc_opt=paths=source_relative \
       helloworld.proto
```

## Usage

### 1. Install dependencies

```bash
go mod tidy
```

### 2. Start the server

```bash
go run server.go
```

### 3. Run the client

In another terminal:

```bash
# Default greeting
go run client.go

# Custom greeting
go run client.go Alice
```

## Expected Output

**Server output:**

```
2024/08/19 10:00:00 server listening at [::]:50051
2024/08/19 10:00:01 Received: world
```

**Client output:**

```
2024/08/19 10:00:01 Greeting: Hello world
```
