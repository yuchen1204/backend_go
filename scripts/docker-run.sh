#!/bin/bash

echo "🚀 启动Backend容器..."

# 停止并删除已存在的容器
docker stop backend-container 2>/dev/null
docker rm backend-container 2>/dev/null

# 创建上传目录（如果不存在）
mkdir -p uploads/{default,avatar,document}

# 启动新容器
docker run -d \
    --name backend-container \
    -p 8080:8080 \
    -p 2222:22 \
    -p 5432:5432 \
    -p 6379:6379 \
    -v $(pwd)/uploads:/app/uploads \
    -v $(pwd)/configs:/app/configs \
    -e JWT_SECRET=your-super-secret-jwt-key-change-this-in-production \
    backend-app:latest

if [ $? -eq 0 ]; then
    echo "✅ 容器启动成功！"
    echo ""
    echo "📋 服务访问信息:"
    echo "  🌐 Web应用:    http://localhost:8080"
    echo "  📚 API文档:    http://localhost:8080/swagger/index.html"
    echo "  🔧 SSH访问:    ssh root@localhost -p 2222 (密码: root)"
    echo "  🗄️ PostgreSQL: localhost:5432 (用户: postgres, 密码: postgres)"
    echo "  🚀 Redis:      localhost:6379"
    echo ""
    echo "🔍 实用命令:"
    echo "  📊 查看日志:   docker logs backend-container"
    echo "  💻 进入容器:   docker exec -it backend-container bash"
    echo "  ⏹️ 停止容器:   docker stop backend-container"
    echo "  🗑️ 删除容器:   docker rm backend-container"
    echo ""
    echo "⏳ 等待服务启动完成..."
    sleep 5
    
    echo "🔍 检查服务状态:"
    docker exec backend-container supervisorctl status
else
    echo "❌ 容器启动失败！"
    exit 1
fi 