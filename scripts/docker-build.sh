#!/bin/bash

echo "🐳 构建Backend Docker镜像..."

# 检查是否存在docker目录
if [ ! -d "docker" ]; then
    echo "❌ docker目录不存在！"
    exit 1
fi

# 构建镜像
docker build -t backend-app:latest .

if [ $? -eq 0 ]; then
    echo "✅ Docker镜像构建成功！"
    echo "💡 镜像名称: backend-app:latest"
    echo ""
    echo "📋 可用命令:"
    echo "  🚀 启动容器: ./scripts/docker-run.sh"
    echo "  🐙 Docker Compose: docker-compose up -d"
    echo "  📊 查看镜像: docker images | grep backend-app"
else
    echo "❌ Docker镜像构建失败！"
    exit 1
fi 