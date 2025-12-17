#!/bin/bash

# HTTP服务性能基准测试
# 使用ab (Apache Bench) 测试 /nextid 端点

SERVER_URL="${SERVER_URL:-http://localhost:8080}"
ENDPOINT="${ENDPOINT:-/nextid}"

echo "=== HTTP服务性能基准测试 ==="
echo "服务器: $SERVER_URL"
echo "端点: $ENDPOINT"
echo ""

# 检查服务器是否可用
echo "检查服务器是否可用..."
if ! curl -s -f "$SERVER_URL/health" > /dev/null 2>&1; then
    echo "错误: 服务器不可用，请先启动HTTP服务器"
    echo ""
    echo "启动命令:"
    echo "  go run example_http/main.go -addr :8080 -redis localhost:6379"
    echo ""
    echo "或设置环境变量:"
    echo "  export REDIS_ADDR=localhost:6379"
    echo "  go run example_http/main.go"
    exit 1
fi
echo "✓ 服务器可用"
echo ""

# 测试单个请求
echo "=== 测试单个请求 ==="
RESPONSE=$(curl -s "$SERVER_URL$ENDPOINT")
echo "响应: $RESPONSE"
echo ""

# 检查是否有ab工具
if ! command -v ab > /dev/null 2>&1; then
    echo "错误: 未找到ab工具"
    echo ""
    echo "安装方法:"
    echo "  macOS: brew install httpd"
    echo "  Ubuntu/Debian: sudo apt-get install apache2-utils"
    echo "  CentOS/RHEL: sudo yum install httpd-tools"
    exit 1
fi

echo "=== 性能测试配置 ==="
echo "测试1: 100请求，并发10"
echo "测试2: 1000请求，并发50"
echo "测试3: 10000请求，并发100"
echo ""

# 测试1: 小规模测试
echo "--- 测试1: 100请求，并发10 ---"
ab -n 100 -c 10 -q "$SERVER_URL$ENDPOINT" | grep -E "Requests per second|Time per request|Transfer rate|Failed requests"
echo ""

# 测试2: 中等规模测试
echo "--- 测试2: 1000请求，并发50 ---"
ab -n 1000 -c 50 -q "$SERVER_URL$ENDPOINT" | grep -E "Requests per second|Time per request|Transfer rate|Failed requests"
echo ""

# 测试3: 大规模测试
echo "--- 测试3: 10000请求，并发100 ---"
ab -n 10000 -c 100 -q "$SERVER_URL$ENDPOINT" | grep -E "Requests per second|Time per request|Transfer rate|Failed requests"
echo ""

echo "=== 测试完成 ==="

