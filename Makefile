.PHONY: all build clean run-proxy run-server run-client run test stop restart proto help

# 默认目标：编译所有组件
all: build

# 编译所有组件
build: build-proxy build-server build-client

# 编译代理服务器
build-proxy:
	@echo "编译代理服务器..."
	@go build -o grpc_proxy main.go

# 编译示例服务器
build-server:
	@echo "编译示例服务器..."
	@cd examples/server && go build -o server main.go

# 编译示例客户端
build-client:
	@echo "编译示例客户端..."
	@cd examples/client && go build -o client main.go

# 清理编译产物
clean:
	@echo "清理编译产物..."
	@rm -f grpc_proxy
	@rm -f examples/server/server
	@rm -f examples/client/client

# 运行代理服务器（后台）
run-proxy:
	@echo "启动代理服务器..."
	@./grpc_proxy --listen=:50051 --target=:50052 &
	@echo "代理服务器已启动，监听端口 50051，转发到 50052"

# 运行示例服务器（后台）
run-server:
	@echo "启动示例服务器..."
	@./examples/server/server --addr=:50052 &
	@echo "示例服务器已启动，监听端口 50052"

# 运行示例客户端
run-client:
	@echo "运行示例客户端..."
	@./examples/client/client --addr=:50051 --name="测试用户"

# 启动所有服务
run: run-server run-proxy
	@echo "所有服务已启动"

# 运行完整测试
test: run
	@echo "运行测试..."
	@sleep 2
	@make run-client

# 停止所有服务
stop:
	@echo "停止所有服务..."
	@-pkill -f "grpc_proxy"
	@-pkill -f "examples/server/server"
	@echo "所有服务已停止"

# 重启所有服务
restart: stop run

# 编译proto文件
proto:
	@echo "检查protoc编译器..."
	@which protoc > /dev/null || (echo "错误: protoc未安装，请先安装protoc" && exit 1)
	@echo "编译proto文件..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/hello.proto
	@echo "Proto文件编译成功"

# 显示帮助信息
help:
	@echo "可用命令："
	@echo "  make build      - 编译所有组件"
	@echo "  make clean      - 清理编译产物"
	@echo "  make run        - 启动服务器和代理"
	@echo "  make run-client - 运行客户端"
	@echo "  make test       - 运行完整测试"
	@echo "  make stop       - 停止所有服务"
	@echo "  make restart    - 重启所有服务"
	@echo "  make proto      - 重新编译proto文件" 