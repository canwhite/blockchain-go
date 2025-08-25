package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure" 是 Go gRPC 官方库提供的一个包，
	// 用于创建不安全（即不加密、不验证身份）的 gRPC 连接凭证（credentials）。
	"google.golang.org/grpc/credentials/insecure"
	"grpc-simple/helloworld"
)

const port = ":50051"

type server struct {
	helloworld.UnimplementedGreeterServer
}

//0）接收者声明，实现say hello 方法
func (s *server) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &helloworld.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func startServer() {
	//1)创建tcp监听
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	//2）创建grpc 服务器实例
	s := grpc.NewServer()
	//3）注册服务实现
	helloworld.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	//4）启动服务监听
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func runClient() {
	//address
	address := "localhost:50051"
	defaultName := "world"
	//dial是拨号的意思
	// 这三个参数分别是：
	// 1. address：服务端的地址（如 "localhost:50051"），客户端需要知道服务端监听的主机和端口才能连接。
	// 2. grpc.WithTransportCredentials(insecure.NewCredentials())：指定连接的安全凭证。这里用的是 insecure.NewCredentials()，表示不加密、不验证身份（明文传输），适合开发测试环境。
	// 3. grpc.WithBlock()：让 Dial 操作变为阻塞模式，直到连接建立或出错才返回。否则 Dial 默认是异步的，可能还没连上就返回了。
	//
	// gRPC 是基于 HTTP/2 协议的长连接通信。客户端通过 grpc.Dial 和服务端建立 TCP 连接后，会一直保持这个连接（除非手动关闭或发生异常）。
	// 在这个连接上，客户端可以多次发起 RPC 调用，服务端也可以推送流式数据。HTTP/2 支持多路复用，多个请求可以复用同一个连接，提升效率和实时性。

	//1）建立与服务器的连接，使用的是不安全的连接方式
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	//2）创建客户端实例
	c := helloworld.NewGreeterClient(conn)

	//3）调用sayHello方法
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	
	// time.Second 是 Go 语言 time 包中定义的一个常量，表示 1 秒的时间间隔（类型为 time.Duration）。
	// 例如 context.WithTimeout(context.Background(), time.Second) 表示设置 1 秒的超时时间。
	// 你也可以用 2*time.Second 表示 2 秒，500*time.Millisecond 表示 0.5 秒。
	ctx, cancel := context.WithTimeout(context.Background(),3 * time.Second)
	defer cancel()
	//默认是让人嵌入ctx的
	r, err := c.SayHello(ctx, &helloworld.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "client" {
		runClient()
	} else {
		startServer()
	}
}