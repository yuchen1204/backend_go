#!/bin/bash

# 创建日志目录
mkdir -p /var/log/supervisor

# 启动PostgreSQL进行初始化
echo "启动PostgreSQL..."
service postgresql start
sleep 5

# 检查数据库是否存在，如果不存在则创建
su - postgres -c "psql -lqt | cut -d \| -f 1 | grep -qw backend || createdb backend"

# 确保PostgreSQL用户和权限正确
su - postgres -c "psql -c \"ALTER USER postgres PASSWORD 'postgres';\""

# 停止PostgreSQL（稍后由supervisor管理）
service postgresql stop

# 启动Redis进行测试
echo "测试Redis..."
service redis-server start
sleep 2
redis-cli ping
service redis-server stop

# 生成SSH主机密钥（如果不存在）
if [ ! -f /etc/ssh/ssh_host_rsa_key ]; then
    ssh-keygen -t rsa -f /etc/ssh/ssh_host_rsa_key -N ''
fi
if [ ! -f /etc/ssh/ssh_host_ecdsa_key ]; then
    ssh-keygen -t ecdsa -f /etc/ssh/ssh_host_ecdsa_key -N ''
fi
if [ ! -f /etc/ssh/ssh_host_ed25519_key ]; then
    ssh-keygen -t ed25519 -f /etc/ssh/ssh_host_ed25519_key -N ''
fi

# 设置权限
chown -R postgres:postgres /var/lib/postgresql
chmod 700 /var/lib/postgresql/13/main

# 创建.env文件（如果不存在）
if [ ! -f /app/.env ]; then
    cp /app/configs/env.example /app/.env
    
    # 更新.env文件中的数据库配置
    sed -i 's/DB_HOST=localhost/DB_HOST=localhost/' /app/.env
    sed -i 's/DB_PASSWORD=postgres/DB_PASSWORD=postgres/' /app/.env
    sed -i 's/REDIS_HOST=localhost/REDIS_HOST=localhost/' /app/.env
fi

echo "所有服务初始化完成，启动supervisor..."

# 启动supervisor管理所有服务
exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf 