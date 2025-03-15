package config

// Config 包含代理服务器的配置选项
type Config struct {
	// 代理服务器监听地址
	ListenAddr string
	// 目标服务器地址
	TargetAddr string
	// 是否启用TLS
	EnableTLS bool
	// TLS证书文件路径
	CertFile string
	// TLS密钥文件路径
	KeyFile string
	// 是否启用日志
	EnableLogging bool
	// 最大并发连接数
	MaxConcurrentStreams uint32
	// 连接超时时间（秒）
	ConnectionTimeout int
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ListenAddr:           ":50051",
		TargetAddr:           ":50052",
		EnableTLS:            false,
		EnableLogging:        true,
		MaxConcurrentStreams: 100,
		ConnectionTimeout:    10,
	}
}
