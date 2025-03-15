#!/bin/bash

# 设置颜色
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # 无颜色

# 显示帮助信息
function show_help {
    echo -e "${YELLOW}gRPC代理构建和运行脚本${NC}"
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  build       - 编译所有组件"
    echo "  clean       - 清理编译产物"
    echo "  run         - 启动服务器和代理"
    echo "  run-client  - 运行客户端"
    echo "  test        - 运行完整测试"
    echo "  stop        - 停止所有服务"
    echo "  restart     - 重启所有服务"
    echo "  proto       - 重新编译proto文件"
    echo "  help        - 显示此帮助信息"
    echo ""
    echo "如果不提供命令，默认执行 'build'"
}

# 编译代理服务器
function build_proxy {
    echo -e "${GREEN}编译代理服务器...${NC}"
    go build -o grpc_proxy main.go
}

# 编译示例服务器
function build_server {
    echo -e "${GREEN}编译示例服务器...${NC}"
    cd examples/server && go build -o server main.go
    cd ../..
}

# 编译示例客户端
function build_client {
    echo -e "${GREEN}编译示例客户端...${NC}"
    cd examples/client && go build -o client main.go
    cd ../..
}

# 编译所有组件
function build_all {
    build_proxy
    build_server
    build_client
    echo -e "${GREEN}所有组件编译完成${NC}"
}

# 清理编译产物
function clean {
    echo -e "${GREEN}清理编译产物...${NC}"
    rm -f grpc_proxy
    rm -f examples/server/server
    rm -f examples/client/client
    echo -e "${GREEN}清理完成${NC}"
}

# 运行代理服务器
function run_proxy {
    echo -e "${GREEN}启动代理服务器...${NC}"
    ./grpc_proxy --listen=:50051 --target=:50052 &
    echo -e "${GREEN}代理服务器已启动，监听端口 50051，转发到 50052${NC}"
}

# 运行示例服务器
function run_server {
    echo -e "${GREEN}启动示例服务器...${NC}"
    ./examples/server/server --addr=:50052 &
    echo -e "${GREEN}示例服务器已启动，监听端口 50052${NC}"
}

# 运行示例客户端
function run_client {
    echo -e "${GREEN}运行示例客户端...${NC}"
    ./examples/client/client --addr=:50051 --name="测试用户"
}

# 启动所有服务
function run_all {
    run_server
    sleep 1
    run_proxy
    echo -e "${GREEN}所有服务已启动${NC}"
}

# 运行完整测试
function run_test {
    run_all
    echo -e "${GREEN}运行测试...${NC}"
    sleep 2
    run_client
}

# 停止所有服务
function stop_all {
    echo -e "${GREEN}停止所有服务...${NC}"
    pkill -f "grpc_proxy" 2>/dev/null || true
    pkill -f "examples/server/server" 2>/dev/null || true
    echo -e "${GREEN}所有服务已停止${NC}"
}

# 重启所有服务
function restart {
    stop_all
    sleep 1
    run_all
}

# 编译proto文件
function build_proto {
    echo -e "${GREEN}检查protoc编译器和插件...${NC}"
    
    # 检查protoc是否安装
    if ! command -v protoc &> /dev/null; then
        echo -e "${YELLOW}警告: protoc未安装${NC}"
        echo "请安装protoc编译器:"
        echo "  macOS: brew install protobuf"
        echo "  Ubuntu: sudo apt-get install protobuf-compiler"
        echo "  其他系统请参考: https://grpc.io/docs/protoc-installation/"
        exit 1
    fi
    
    # 检查Go插件是否安装
    if ! command -v protoc-gen-go &> /dev/null; then
        echo -e "${YELLOW}安装Go插件...${NC}"
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    fi
    
    echo -e "${GREEN}编译proto文件...${NC}"
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           proto/hello.proto
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Proto文件编译成功${NC}"
    else
        echo -e "${YELLOW}Proto文件编译失败${NC}"
        exit 1
    fi
}

# 主函数
function main {
    case "$1" in
        build)
            build_all
            ;;
        clean)
            clean
            ;;
        run)
            run_all
            ;;
        run-client)
            run_client
            ;;
        test)
            run_test
            ;;
        stop)
            stop_all
            ;;
        restart)
            restart
            ;;
        proto)
            build_proto
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            # 默认执行build
            if [ -z "$1" ]; then
                build_all
            else
                echo -e "${YELLOW}未知命令: $1${NC}"
                show_help
                exit 1
            fi
            ;;
    esac
}

# 执行主函数
main "$@" 