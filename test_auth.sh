#!/bin/bash

echo "=== 测试文件上传认证功能 ==="

# 服务器地址
BASE_URL="http://localhost:8080/api/v1"

# 1. 注册测试用户
echo "1. 注册测试用户..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/users/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser123",
    "email": "test123@example.com",
    "password": "password123",
    "verification_code": "000000",
    "nickname": "测试用户"
  }')

echo "注册响应: $REGISTER_RESPONSE"

# 2. 登录获取token
echo -e "\n2. 登录获取token..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/users/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser123",
    "password": "password123"
  }')

echo "登录响应: $LOGIN_RESPONSE"

# 提取access_token
ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
    echo "❌ 无法获取access_token，登录失败"
    exit 1
fi

echo "✅ 获取到access_token: ${ACCESS_TOKEN:0:20}..."

# 3. 测试文件列表接口（这是报告问题的接口）
echo -e "\n3. 测试获取用户文件列表..."
FILES_RESPONSE=$(curl -s -X GET "$BASE_URL/files/my" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "文件列表响应: $FILES_RESPONSE"

# 检查是否还返回401错误
if echo "$FILES_RESPONSE" | grep -q "未授权\|401"; then
    echo "❌ 认证问题仍然存在！"
    exit 1
else
    echo "✅ 认证问题已解决！"
fi

# 4. 测试存储信息接口
echo -e "\n4. 测试获取存储信息..."
STORAGE_RESPONSE=$(curl -s -X GET "$BASE_URL/files/storages")
echo "存储信息响应: $STORAGE_RESPONSE"

echo -e "\n=== 测试完成 ===" 