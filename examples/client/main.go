package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	pb "github.com/jiwubu/grpc_proxy/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 服务器地址
var addr = flag.String("addr", ":50051", "服务器地址")
var name = flag.String("name", "世界", "要问候的名称")

func main() {
	flag.Parse()

	// 连接到服务器
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := pb.NewHelloServiceClient(conn)

	// 设置上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 调用一元RPC
	callUnary(ctx, client)

	// 调用服务器流式RPC
	log.Printf("--------------------------------")
	callServerStream(ctx, client)

	// 调用客户端流式RPC
	log.Printf("--------------------------------")
	callClientStream(ctx, client)

	// 调用双向流式RPC
	log.Printf("--------------------------------")
	callBidirectionalStream(ctx, client)
}

// 调用一元RPC
func callUnary(ctx context.Context, client pb.HelloServiceClient) {
	log.Println("调用一元RPC...")
	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("一元RPC调用失败: %v", err)
	}
	log.Printf("一元RPC响应: %s", resp.GetMessage())
}

// 调用服务器流式RPC
func callServerStream(ctx context.Context, client pb.HelloServiceClient) {
	log.Println("调用服务器流式RPC...")
	stream, err := client.SayHelloServerStream(ctx, &pb.HelloRequest{Name: "服务器流测试"})
	if err != nil {
		log.Fatalf("服务器流式RPC调用失败: %v", err)
	}

	// 接收流
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("接收服务器流失败: %v", err)
		}
		log.Printf("服务器流响应: %s", resp.GetMessage())
	}
}

// 调用客户端流式RPC
func callClientStream(ctx context.Context, client pb.HelloServiceClient) {
	log.Println("调用客户端流式RPC...")
	stream, err := client.SayHelloClientStream(ctx)
	if err != nil {
		log.Fatalf("客户端流式RPC调用失败: %v", err)
	}

	// 发送流
	names := []string{"客户端", "流", "测试"}
	for _, name := range names {
		if err := stream.Send(&pb.HelloRequest{Name: name}); err != nil {
			log.Fatalf("发送客户端流失败: %v", err)
		}
		time.Sleep(time.Second)
	}

	// 关闭并接收响应
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("关闭客户端流失败: %v", err)
	}
	log.Printf("客户端流响应: %s", resp.GetMessage())
}

// 调用双向流式RPC
func callBidirectionalStream(ctx context.Context, client pb.HelloServiceClient) {
	log.Println("调用双向流式RPC...")

	// 创建一个更长的超时上下文
	streamCtx, streamCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer streamCancel()

	stream, err := client.SayHelloBidirectionalStream(streamCtx)
	if err != nil {
		log.Fatalf("双向流式RPC调用失败: %v", err)
	}

	// 创建通道
	waitc := make(chan struct{})
	errChan := make(chan error, 1)

	// 接收goroutine
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				errChan <- err
				close(waitc)
				return
			}
			log.Printf("双向流响应: %s", resp.GetMessage())
		}
	}()

	// 发送请求
	names := []string{"双向", "流", "测试"}
	for _, name := range names {
		if err := stream.Send(&pb.HelloRequest{Name: name}); err != nil {
			log.Fatalf("发送双向流失败: %v", err)
		}
		time.Sleep(time.Second)
	}

	// 关闭发送
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("关闭双向流发送失败: %v", err)
	}

	// 等待接收完成或出错
	select {
	case <-waitc:
		log.Println("双向流RPC完成")
	case err := <-errChan:
		if err != nil {
			log.Printf("接收双向流失败: %v", err)
		}
	case <-time.After(10 * time.Second):
		log.Println("双向流RPC超时，但这是预期的行为")
	}
}
