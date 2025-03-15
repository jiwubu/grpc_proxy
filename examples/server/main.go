package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	pb "github.com/jiwubu/grpc_proxy/proto"
	"google.golang.org/grpc"
)

// 服务器地址
var addr = flag.String("addr", ":50052", "服务器地址")

// HelloServer 实现Hello服务
type HelloServer struct {
	pb.UnimplementedHelloServiceServer
}

// SayHello 实现一元RPC
func (s *HelloServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("收到一元请求: %v", req.GetName())
	return &pb.HelloResponse{Message: fmt.Sprintf("你好, %s!", req.GetName())}, nil
}

// SayHelloServerStream 实现服务器流式RPC
func (s *HelloServer) SayHelloServerStream(req *pb.HelloRequest, stream pb.HelloService_SayHelloServerStreamServer) error {
	log.Printf("收到服务器流式请求: %v", req.GetName())

	// 发送5条消息
	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("你好 %s, 消息 #%d", req.GetName(), i+1)
		if err := stream.Send(&pb.HelloResponse{Message: msg}); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}

	return nil
}

// SayHelloClientStream 实现客户端流式RPC
func (s *HelloServer) SayHelloClientStream(stream pb.HelloService_SayHelloClientStreamServer) error {
	var names []string

	// 接收客户端流
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// 流结束
			break
		}
		if err != nil {
			return err
		}

		log.Printf("收到客户端流消息: %v", req.GetName())
		names = append(names, req.GetName())
	}

	// 发送响应
	response := fmt.Sprintf("你好: %v!", names)
	return stream.SendAndClose(&pb.HelloResponse{Message: response})
}

// SayHelloBidirectionalStream 实现双向流式RPC
func (s *HelloServer) SayHelloBidirectionalStream(stream pb.HelloService_SayHelloBidirectionalStreamServer) error {
	// 接收并发送
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		log.Printf("收到双向流请求: %v", req.GetName())

		// 发送响应
		response := fmt.Sprintf("你好, %s!", req.GetName())
		if err := stream.Send(&pb.HelloResponse{Message: response}); err != nil {
			return err
		}
	}
}

func main() {
	flag.Parse()

	// 创建监听器
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	// 创建gRPC服务器
	s := grpc.NewServer()

	// 注册服务
	pb.RegisterHelloServiceServer(s, &HelloServer{})

	log.Printf("服务器启动在 %s", *addr)

	// 启动服务器
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务失败: %v", err)
	}
}
