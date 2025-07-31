# 多阶段构建，基于Ubuntu 22.04
FROM ubuntu:22.04

# 设置环境变量
ENV DEBIAN_FRONTEND=noninteractive
ENV GO_VERSION=1.24.5
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# 创建工作目录
WORKDIR /app

# 安装基础工具和依赖
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    supervisor \
    openssh-server \
    postgresql \
    postgresql-contrib \
    redis-server \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# 安装Go语言
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

# 配置PostgreSQL
RUN service postgresql start && \
    sudo -u postgres psql -c "CREATE USER postgres WITH PASSWORD 'postgres';" && \
    sudo -u postgres psql -c "ALTER USER postgres CREATEDB;" && \
    sudo -u postgres createdb backend && \
    service postgresql stop

# 配置Redis
RUN sed -i 's/bind 127.0.0.1 ::1/bind 0.0.0.0/' /etc/redis/redis.conf && \
    sed -i 's/protected-mode yes/protected-mode no/' /etc/redis/redis.conf

# 配置SSH
RUN mkdir /var/run/sshd && \
    echo 'root:password' | chpasswd && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#Port 22/Port 1000/' /etc/ssh/sshd_config && \
    ssh-keygen -A

# 复制项目文件
COPY . /app/

# 设置Go模块代理（加速下载）
ENV GOPROXY=https://goproxy.cn,direct

# 构建Go应用
RUN go mod tidy && \
    go mod download && \
    go build -o backend cmd/main.go

# 创建上传目录
RUN mkdir -p uploads/default uploads/avatar uploads/document

# 创建环境配置文件
RUN cat > .env << 'EOF'
# 服务器配置
PORT=1001

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=backend
DB_SSLMODE=disable

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# SMTP 邮件服务配置（可选）
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-email@example.com
SMTP_PASSWORD=your-email-password
SMTP_FROM=your-email@example.com

# 安全配置
MAX_IP_REQUESTS_PER_DAY=1000

# JWT 配置
JWT_SECRET=your-very-secret-jwt-key-change-this-in-production
JWT_ACCESS_TOKEN_EXPIRES_IN_MINUTES=30
JWT_REFRESH_TOKEN_EXPIRES_IN_DAYS=7

# 文件存储配置
FILE_STORAGE_DEFAULT=local_default

# 本地存储配置
FILE_STORAGE_LOCAL_NAMES=default,avatar,document
FILE_STORAGE_LOCAL_DEFAULT_PATH=./uploads/default
FILE_STORAGE_LOCAL_DEFAULT_URL=http://localhost:1001/uploads/default
FILE_STORAGE_LOCAL_AVATAR_PATH=./uploads/avatar
FILE_STORAGE_LOCAL_AVATAR_URL=http://localhost:1001/uploads/avatar
FILE_STORAGE_LOCAL_DOCUMENT_PATH=./uploads/document
FILE_STORAGE_LOCAL_DOCUMENT_URL=http://localhost:1001/uploads/document
EOF

# 创建supervisor配置文件
RUN cat > /etc/supervisor/conf.d/services.conf << 'EOF'
[supervisord]
nodaemon=true
user=root

[program:postgresql]
command=/usr/lib/postgresql/14/bin/postgres -D /var/lib/postgresql/14/main -c config_file=/etc/postgresql/14/main/postgresql.conf
user=postgres
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/postgresql.log

[program:redis]
command=redis-server /etc/redis/redis.conf
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/redis.log

[program:sshd]
command=/usr/sbin/sshd -D -p 1000
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/sshd.log

[program:backend]
command=/app/backend
directory=/app
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/backend.log
environment=PATH="/usr/local/go/bin:%(ENV_PATH)s"
EOF

# 创建启动脚本
RUN cat > /start.sh << 'EOF'
#!/bin/bash

# 等待PostgreSQL启动
echo "启动PostgreSQL..."
service postgresql start
sleep 5

# 确保数据库和用户存在
sudo -u postgres psql -c "SELECT 1;" 2>/dev/null || {
    echo "初始化PostgreSQL..."
    sudo -u postgres psql -c "CREATE USER postgres WITH PASSWORD 'postgres';" 2>/dev/null || true
    sudo -u postgres psql -c "ALTER USER postgres CREATEDB;" 2>/dev/null || true
    sudo -u postgres createdb backend 2>/dev/null || true
}

# 停止PostgreSQL（由supervisor管理）
service postgresql stop
sleep 2

# 启动supervisor管理所有服务
echo "启动所有服务..."
exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf
EOF

RUN chmod +x /start.sh

# 暴露端口
EXPOSE 1000 1001

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:1001/api/v1/health || exit 1

# 设置数据卷
VOLUME ["/var/lib/postgresql", "/var/lib/redis", "/app/uploads"]

# 启动脚本
CMD ["/start.sh"] 