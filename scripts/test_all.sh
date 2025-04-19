#!/bin/bash

# 定义颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== 启动测试环境 ===${NC}"

# 确保运行目录正确
cd "$(dirname "$0")/.."

# 1. 启动测试服务器
echo "启动后端测试服务器..."
go run test/test_servers.go &
TEST_SERVERS_PID=$!

# 等待服务器启动
sleep 2

# 2. 启动网关服务
echo "启动网关服务..."
go run cmd/server/main.go &
GATEWAY_PID=$!

# 等待网关启动
sleep 2

# 清理函数
cleanup() {
    echo -e "\n${YELLOW}=== 清理测试环境 ===${NC}"
    echo "停止网关服务 (PID: $GATEWAY_PID)..."
    kill $GATEWAY_PID 2>/dev/null
    echo "停止测试服务器 (PID: $TEST_SERVERS_PID)..."
    kill $TEST_SERVERS_PID 2>/dev/null
    exit
}

# 捕获Ctrl+C和退出信号
trap cleanup EXIT INT TERM

echo -e "\n${YELLOW}=== 测试JWT鉴权 ===${NC}"

# 测试健康检查 (不需要鉴权)
echo "1. 测试健康检查 (公开API)..."
health_response=$(curl -s http://localhost:8080/health)
if [[ $health_response == *"ok"* ]]; then
    echo -e "${GREEN}✓ 健康检查成功${NC}"
else
    echo -e "${RED}✗ 健康检查失败: $health_response${NC}"
fi

# 测试未授权访问
echo -e "\n2. 测试未授权访问..."
unauth_response=$(curl -s http://localhost:8080/api/test)
if [[ $unauth_response == *"authorization header is required"* ]]; then
    echo -e "${GREEN}✓ 未授权访问被正确拒绝${NC}"
else
    echo -e "${RED}✗ 未授权访问测试失败: $unauth_response${NC}"
fi

# 获取JWT令牌
echo -e "\n3. 获取JWT令牌..."
token_response=$(curl -s -X POST http://localhost:8080/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"test"}')
token=$(echo $token_response | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -n "$token" ]; then
    echo -e "${GREEN}✓ 成功获取令牌${NC}"
    # 显示令牌的部分内容
    echo "   令牌: ${token:0:20}..."
else
    echo -e "${RED}✗ 获取令牌失败: $token_response${NC}"
    exit 1
fi

echo -e "\n${YELLOW}=== 测试负载均衡和反向代理 ===${NC}"

# 测试带JWT的API访问
echo "4. 测试带JWT的API访问和负载均衡..."

# 使用JWT访问API多次，观察负载均衡效果
server1_count=0
server2_count=0
total_requests=20

echo "   发送 $total_requests 个请求并观察负载均衡效果..."
for i in $(seq 1 $total_requests); do
    response=$(curl -s http://localhost:8080/api/test \
        -H "Authorization: Bearer $token")
    
    # 检查请求是否成功    
    if [[ $response == *"Hello from test server"* ]]; then
        echo -n "."
        
        # 统计服务器分布
        if [[ $response == *"1 (Weight: 3)"* ]]; then
            ((server1_count++))
        elif [[ $response == *"2 (Weight: 2)"* ]]; then
            ((server2_count++))
        fi
    else
        echo -e "\n${RED}✗ 请求失败: $response${NC}"
    fi
done

echo -e "\n   请求分布统计:"
echo "   - 服务器1 (权重3): $server1_count 请求 ($(echo "scale=1; $server1_count*100/$total_requests" | bc)%)"
echo "   - 服务器2 (权重2): $server2_count 请求 ($(echo "scale=1; $server2_count*100/$total_requests" | bc)%)"

# 检查负载均衡是否符合权重比例 (约3:2)
if [ $server1_count -gt $server2_count ]; then
    echo -e "${GREEN}✓ 负载均衡符合权重设置 (服务器1 > 服务器2)${NC}"
else
    echo -e "${YELLOW}! 负载均衡分布可能不符合权重设置${NC}"
fi

# 测试路径重写
echo -e "\n5. 测试路径重写..."
# 检查响应是否来自后端的 /test 路径 (没有 /api 前缀)
if [[ $response == *"Hello from test server"* ]]; then
    echo -e "${GREEN}✓ 路径重写成功 (/api/test -> /test)${NC}"
else
    echo -e "${RED}✗ 路径重写测试失败${NC}"
fi

echo -e "\n${GREEN}=== 测试完成 ===${NC}" 