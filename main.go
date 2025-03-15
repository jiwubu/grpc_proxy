package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jiwubu/grpc_proxy/config"
	"github.com/jiwubu/grpc_proxy/proxy"
)

func main() {
	// 解析命令行参数
	listenAddr := flag.String("listen", ":50051", "代理服务器监听地址")
	targetAddr := flag.String("target", ":50052", "目标服务器地址")
	enableTLS := flag.Bool("tls", false, "是否启用TLS")
	certFile := flag.String("cert", "", "TLS证书文件路径")
	keyFile := flag.String("key", "", "TLS密钥文件路径")
	enableLogging := flag.Bool("log", true, "是否启用日志")
	maxStreams := flag.Uint("max-streams", 100, "最大并发流数量")
	timeout := flag.Int("timeout", 10, "连接超时时间（秒）")

	flag.Parse()

	// 创建配置
	cfg := &config.Config{
		ListenAddr:           *listenAddr,
		TargetAddr:           *targetAddr,
		EnableTLS:            *enableTLS,
		CertFile:             *certFile,
		KeyFile:              *keyFile,
		EnableLogging:        *enableLogging,
		MaxConcurrentStreams: uint32(*maxStreams),
		ConnectionTimeout:    *timeout,
	}

	// 创建代理
	p := proxy.NewGRPCProxy(cfg)

	// 设置信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 启动代理服务器（在goroutine中）
	errCh := make(chan error, 1)
	go func() {
		errCh <- p.Start()
	}()

	// 等待信号或错误
	select {
	case err := <-errCh:
		if err != nil {
			log.Fatalf("代理服务器错误: %v", err)
		}
	case sig := <-sigCh:
		log.Printf("收到信号: %v, 正在关闭服务器...", sig)
		p.Stop()
	}
}
