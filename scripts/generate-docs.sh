#!/bin/bash

echo "正在生成Swagger文档..."

# 检查swag工具是否安装
if ! command -v swag &> /dev/null; then
    echo "安装swag工具..."
    /usr/bin/go/bin/go install github.com/swaggo/swag/cmd/swag@latest
fi

# 生成Swagger文档
echo "正在生成API文档..."
swag init -g cmd/main.go -o ./docs

if [ $? -eq 0 ]; then
    echo "✅ Swagger文档生成成功!"
    echo "📚 文档位置: ./docs/"
    echo "🌐 启动服务后访问: http://localhost:1101/swagger/index.html"
else
    echo "❌ Swagger文档生成失败"
    exit 1
fi 