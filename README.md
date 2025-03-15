# gRPC 代理服务器

这是一个用Go语言实现的gRPC代理服务器，可以将gRPC请求从一个端点透明地转发到另一个端点。

## 功能特性

- 支持所有类型的gRPC调用（一元RPC、服务器流式RPC、客户端流式RPC、双向流式RPC）
- 支持TLS加密通信
- 可配置的日志记录
- 可配置的连接超时
- 可配置的最大并发流数量
- 支持优雅关闭
- 支持服务反射

## 安装

### 克隆仓库

```bash
git clone https://github.com/jiwubu/grpc_proxy.git
cd grpc_proxy
```

### 安装依赖

```bash
go mod download
```

## 编译和运行

### 使用脚本（推荐）

项目提供了两种方式来构建和运行：

#### 1. 使用 Shell 脚本

```bash
# 添加执行权限
chmod +x build.sh

# 显示帮助信息
./build.sh help

# 编译所有组件
./build.sh build

# 运行服务
./build.sh run

# 运行客户端
./build.sh run-client

# 运行完整测试
./build.sh test

# 停止所有服务
./build.sh stop
```

#### 2. 使用 Makefile

```bash
# 显示帮助信息
make help

# 编译所有组件
make build

# 运行服务
make run

# 运行客户端
make run-client

# 运行完整测试
make test

# 停止所有服务
make stop
```

### 手动编译和运行

如果您不想使用脚本，也可以手动编译和运行：

```bash
# 编译代理服务器
go build -o grpc_proxy main.go

# 编译示例服务器
cd examples/server && go build -o server main.go && cd ../..

# 编译示例客户端
cd examples/client && go build -o client main.go && cd ../..

# 运行示例服务器
./examples/server/server --addr=:50052 &

# 运行代理服务器
./grpc_proxy --listen=:50051 --target=:50052 &

# 运行示例客户端
./examples/client/client --addr=:50051 --name="测试用户"
```

## 编译 Proto 文件

如果您修改了 proto 定义文件，需要重新生成 Go 代码：

```bash
# 安装 protoc 编译器（如果尚未安装）
# macOS
brew install protobuf

# Ubuntu
sudo apt-get install protobuf-compiler

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 编译 proto 文件
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/hello.proto
```

您也可以使用脚本来编译 proto 文件：

```bash
# 使用 Shell 脚本
./build.sh proto

# 使用 Makefile
make proto
```

## 命令行参数

### 代理服务器参数

- `--listen`: 代理服务器监听地址，默认为`:50051`
- `--target`: 目标服务器地址，默认为`:50052`
- `--tls`: 是否启用TLS，默认为`false`
- `--cert`: TLS证书文件路径
- `--key`: TLS密钥文件路径
- `--log`: 是否启用日志，默认为`true`
- `--max-streams`: 最大并发流数量，默认为`100`
- `--timeout`: 连接超时时间（秒），默认为`10`

### 示例服务器参数

- `--addr`: 服务器监听地址，默认为`:50052`

### 示例客户端参数

- `--addr`: 连接地址（代理地址），默认为`:50051`
- `--name`: 发送的名称，默认为`World`

## 示例服务

项目包含了一个示例 HelloService 服务，提供以下 RPC 方法：

- `SayHello`: 一元 RPC
- `SayHelloServerStream`: 服务器流式 RPC
- `SayHelloClientStream`: 客户端流式 RPC
- `SayHelloBidirectionalStream`: 双向流式 RPC

## 架构

该代理服务器使用 gRPC 的拦截器机制来捕获所有传入的请求，并将它们转发到目标服务器。它通过以下步骤工作：

1. 代理服务器监听指定端口
2. 客户端连接到代理服务器
3. 代理服务器拦截所有 RPC 调用
4. 代理服务器将请求转发到目标服务器
5. 代理服务器将目标服务器的响应返回给客户端

## 贡献

欢迎提交问题和拉取请求！

## 许可证

MIT 