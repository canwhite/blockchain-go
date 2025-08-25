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

## 实现总结 (Server vs Client)

### Server (服务端)

**功能**: 启动 gRPC 服务器，监听客户端的问候请求

**主要步骤**:

1. 创建 TCP 监听器 (`net.Listen`) 在端口 50051
2. 创建 gRPC 服务器实例 (`grpc.NewServer()`)
3. 注册 Greeter 服务实现 (`helloworld.RegisterGreeterServer`)
4. 启动服务监听 (`server.Serve`)，阻塞等待客户端连接
5. 实现`SayHello`方法：接收客户端发送的名字参数，返回"Hello + 名字"的响应

**核心逻辑**: `main.go:23-26` - 实现 gRPC 服务接口，处理客户端请求

### Client (客户端)

**功能**: 连接服务器并发送问候请求

**主要步骤**:

1. 建立到服务器的连接 (`grpc.Dial`)，使用不安全传输模式
2. 创建 Greeter 客户端实例 (`helloworld.NewGreeterClient`)
3. 准备请求数据 (默认"world"或命令行参数)
4. 调用`SayHello`方法发送 gRPC 请求，设置 1 秒超时
5. 打印服务器返回的问候消息

**核心逻辑**: `main.go:77` - 发起 gRPC 调用，获取服务器响应

### 通信流程

```
Client --(TCP连接)--> Server
Client --SayHello(name)--> Server
Server --HelloReply(message)--> Client
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
