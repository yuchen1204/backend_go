#!/bin/bash

echo "📊 Backend容器日志监控"
echo "======================="

# 检查容器是否运行
if ! docker ps | grep -q backend-container; then
    echo "❌ backend-container 容器未运行！"
    echo "💡 请先运行: ./scripts/docker-run.sh"
    exit 1
fi

echo "选择要查看的日志:"
echo "1) 所有日志"
echo "2) Go应用日志" 
echo "3) PostgreSQL日志"
echo "4) Redis日志"
echo "5) SSH日志"
echo "6) Supervisor状态"
echo "7) 实时监控所有日志"

read -p "请选择 (1-7): " choice

case $choice in
    1)
        echo "📋 显示所有日志..."
        docker logs backend-container
        ;;
    2)
        echo "🔍 Go应用日志..."
        docker exec backend-container tail -f /var/log/backend.log
        ;;
    3)
        echo "🗄️ PostgreSQL日志..."
        docker exec backend-container tail -f /var/log/postgresql.log
        ;;
    4)
        echo "🚀 Redis日志..."
        docker exec backend-container tail -f /var/log/redis.log
        ;;
    5)
        echo "🔐 SSH日志..."
        docker exec backend-container tail -f /var/log/sshd.log
        ;;
    6)
        echo "📊 Supervisor服务状态..."
        docker exec backend-container supervisorctl status
        ;;
    7)
        echo "📈 实时监控所有日志..."
        docker logs -f backend-container
        ;;
    *)
        echo "❌ 无效选择！"
        exit 1
        ;;
esac 