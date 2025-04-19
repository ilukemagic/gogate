#!/bin/bash

# 清晰显示每个步骤
set -x

# 先获取token
echo "获取JWT令牌..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}')

echo "令牌响应: $TOKEN_RESPONSE"

# 手动提取token (适用于简单JSON格式)
TOKEN=$(echo $TOKEN_RESPONSE | sed 's/.*"token":"\([^"]*\)".*/\1/')
echo "提取的令牌: $TOKEN"

if [ -z "$TOKEN" ]; then
  echo "无法获取令牌，请检查登录服务"
  exit 1
fi

# 测试限流 - 先测试单个请求是否能成功
echo "测试单个请求..."
TEST_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/test)
echo "测试响应: $TEST_RESPONSE"

# 如果单个请求成功，继续测试限流
if [ $? -eq 0 ]; then
  echo "单个请求成功，开始测试限流..."
  
  # 测试限流
  SUCCESS=0
  LIMITED=0
  TOTAL=30
  
  echo "开始发送 $TOTAL 个快速请求..."
  for i in $(seq 1 $TOTAL); do
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
      -H "Authorization: Bearer $TOKEN" \
      http://localhost:8080/api/test)
    
    echo -n "请求 $i: 状态码=$HTTP_CODE "
    
    if [ "$HTTP_CODE" == "200" ]; then
      SUCCESS=$((SUCCESS+1))
      echo "成功"
    elif [ "$HTTP_CODE" == "429" ]; then
      LIMITED=$((LIMITED+1))
      echo "限流"
    else
      echo "其他状态"
    fi
    
    # 减少延迟使限流更明显
    sleep 0.01
  done
  
  echo
  echo "测试结果: 成功=$SUCCESS, 限流=$LIMITED, 总计=$TOTAL"
  
  if [ "$LIMITED" -eq 0 ]; then
    echo "警告: 没有请求被限流，限流功能可能未生效!"
  else
    echo "限流功能工作正常，检测到 $LIMITED 个被限流的请求"
  fi
else
  echo "单个请求失败，请先确保API可以正常访问"
fi