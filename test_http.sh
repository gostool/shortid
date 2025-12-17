#!/bin/bash

# HTTP服务性能测试脚本
# 使用ab (Apache Bench) 或curl进行性能测试

SERVER_URL="${SERVER_URL:-http://localhost:8080}"
ENDPOINT="${ENDPOINT:-/nextid}"
REQUESTS="${REQUESTS:-1000}"
CONCURRENCY="${CONCURRENCY:-10}"

echo "=== HTTP服务性能测试 ==="
echo "服务器: $SERVER_URL"
echo "端点: $ENDPOINT"
echo "请求数: $REQUESTS"
echo "并发数: $CONCURRENCY"
echo ""

# 检查服务器是否可用
echo "检查服务器是否可用..."
if ! curl -s -f "$SERVER_URL/health" > /dev/null 2>&1; then
    echo "错误: 服务器不可用，请先启动HTTP服务器"
    echo "启动命令: go run example_http/main.go -addr :8080 -redis localhost:6379"
    exit 1
fi
echo "✓ 服务器可用"
echo ""

# 测试单个请求
echo "=== 测试单个请求 ==="
RESPONSE=$(curl -s "$SERVER_URL$ENDPOINT")
echo "响应: $RESPONSE"
echo ""

# 如果有ab工具，使用ab进行性能测试
if command -v ab > /dev/null 2>&1; then
    echo "=== 使用Apache Bench (ab) 进行性能测试 ==="
    ab -n $REQUESTS -c $CONCURRENCY "$SERVER_URL$ENDPOINT"
elif command -v wrk > /dev/null 2>&1; then
    echo "=== 使用wrk进行性能测试 ==="
    wrk -t4 -c$CONCURRENCY -d10s "$SERVER_URL$ENDPOINT"
else
    echo "=== 使用curl进行简单性能测试 ==="
    echo "开始时间: $(date +%s.%N)"
    START_TIME=$(date +%s.%N)
    
    for i in $(seq 1 $REQUESTS); do
        curl -s "$SERVER_URL$ENDPOINT" > /dev/null
        if [ $((i % 100)) -eq 0 ]; then
            echo "已完成 $i 个请求..."
        fi
    done
    
    END_TIME=$(date +%s.%N)
    DURATION=$(echo "$END_TIME - $START_TIME" | bc)
    QPS=$(echo "scale=2; $REQUESTS / $DURATION" | bc)
    
    echo "结束时间: $(date +%s.%N)"
    echo "总耗时: ${DURATION}秒"
    echo "QPS: $QPS"
fi

