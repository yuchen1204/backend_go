#!/bin/bash

echo "🚀 启动Backend容器（带.env配置）..."

# 检查.env文件是否存在
if [ ! -f ".env" ]; then
    echo "⚠️ .env文件不存在，从模板创建..."
    cp configs/env.example .env
    echo "✅ 已创建.env文件，请根据需要修改配置"
fi

# 停止并删除已存在的容器
docker stop backend-container 2>/dev/null
docker rm backend-container 2>/dev/null

# 创建上传目录（如果不存在）
mkdir -p uploads/{default,avatar,document}

# 启动新容器，挂载.env文件
docker run -d \
    --name backend-container \
    -p 8080:8080 \
    -p 2222:22 \
    -p 5432:5432 \
    -p 6379:6379 \
    -v $(pwd)/uploads:/app/uploads \
    -v $(pwd)/configs:/app/configs \
    -v $(pwd)/.env:/app/.env \
    --env-file .env \
    backend-app:latest

if [ $? -eq 0 ]; then
    echo "✅ 容器启动成功！"
    echo ""
    echo "📋 配置信息:"
    echo "  📁 .env文件:   $(pwd)/.env"
    echo "  🔧 配置加载:   通过 --env-file 和 volume 挂载"
    echo ""
    echo "📋 服务访问信息:"
    echo "  🌐 Web应用:    http://localhost:8080"
    echo "  📚 API文档:    http://localhost:8080/swagger/index.html"
    echo "  🔧 SSH访问:    ssh root@localhost -p 2222 (密码: root)"
    echo "  🗄️ PostgreSQL: localhost:5432 (用户: postgres, 密码: postgres)"
    echo "  🚀 Redis:      localhost:6379"
    echo ""
    echo "🔧 环境变量管理:"
    echo "  📝 编辑配置:   vim .env"
    echo "  🔄 重启应用:   docker restart backend-container"
    echo "  🔍 查看配置:   docker exec backend-container env | grep FILE_STORAGE"
    echo ""
    echo "⏳ 等待服务启动完成..."
    sleep 5
    
    echo "🔍 检查服务状态:"
    docker exec backend-container supervisorctl status
else
    echo "❌ 容器启动失败！"
    exit 1
fi 