package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/jiwubu/grpc_proxy/config"
	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// GRPCProxy 是gRPC代理的实现
type GRPCProxy struct {
	config *config.Config
	server *grpc.Server
}

// NewGRPCProxy 创建一个新的gRPC代理
func NewGRPCProxy(cfg *config.Config) *GRPCProxy {
	return &GRPCProxy{
		config: cfg,
	}
}

// Start 启动代理服务器
func (p *GRPCProxy) Start() error {
	// 创建服务器选项
	var opts []grpc.ServerOption

	// 设置最大并发流
	opts = append(opts, grpc.MaxConcurrentStreams(p.config.MaxConcurrentStreams))

	// 如果启用TLS，添加TLS凭证
	if p.config.EnableTLS {
		creds, err := credentials.NewServerTLSFromFile(p.config.CertFile, p.config.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// 添加拦截器
	if p.config.EnableLogging {
		opts = append(opts, grpc.UnaryInterceptor(p.loggingUnaryInterceptor))
		opts = append(opts, grpc.StreamInterceptor(p.loggingStreamInterceptor))
	}

	// 添加未知服务处理程序
	director := p.buildDirector()
	opts = append(opts, grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))

	// 创建gRPC服务器
	p.server = grpc.NewServer(opts...)

	// 注册反射服务，这样客户端可以发现服务
	reflection.Register(p.server)

	// 启动服务器
	log.Printf("Starting gRPC proxy server on %s, forwarding to %s", p.config.ListenAddr, p.config.TargetAddr)
	lis, err := net.Listen("tcp", p.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	return p.server.Serve(lis)
}

// Stop 停止代理服务器
func (p *GRPCProxy) Stop() {
	if p.server != nil {
		p.server.GracefulStop()
		log.Println("gRPC proxy server stopped")
	}
}

// buildDirector 创建一个流导向器，用于将请求转发到目标服务器
func (p *GRPCProxy) buildDirector() proxy.StreamDirector {
	return func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		if p.config.EnableLogging {
			log.Printf("Proxying RPC: %s", fullMethodName)
		}

		// 转发元数据
		md, _ := metadata.FromIncomingContext(ctx)
		outCtx := metadata.NewOutgoingContext(ctx, md.Copy())

		// 创建到目标服务器的连接
		var opts []grpc.DialOption
		if p.config.EnableTLS {
			creds, err := credentials.NewClientTLSFromFile(p.config.CertFile, "")
			if err != nil {
				return nil, nil, status.Errorf(codes.Internal, "failed to load TLS credentials: %v", err)
			}
			opts = append(opts, grpc.WithTransportCredentials(creds))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		// 设置超时
		dialCtx, cancel := context.WithTimeout(ctx, time.Duration(p.config.ConnectionTimeout)*time.Second)
		defer cancel()

		// 连接到目标服务器
		conn, err := grpc.DialContext(dialCtx, p.config.TargetAddr, opts...)
		if err != nil {
			return nil, nil, status.Errorf(codes.Unavailable, "failed to dial target server: %v", err)
		}

		return outCtx, conn, nil
	}
}

// loggingUnaryInterceptor 是一个用于记录一元RPC调用的拦截器
func (p *GRPCProxy) loggingUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	log.Printf("Unary request: method=%s, metadata=%v", info.FullMethod, md)

	resp, err := handler(ctx, req)

	log.Printf("Unary response: method=%s, duration=%s, error=%v", info.FullMethod, time.Since(start), err)
	return resp, err
}

// loggingStreamInterceptor 是一个用于记录流式RPC调用的拦截器
func (p *GRPCProxy) loggingStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ss.Context())
	log.Printf("Stream request: method=%s, metadata=%v", info.FullMethod, md)

	err := handler(srv, ss)

	log.Printf("Stream response: method=%s, duration=%s, error=%v", info.FullMethod, time.Since(start), err)
	return err
}
